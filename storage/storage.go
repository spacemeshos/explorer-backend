package storage

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spacemeshos/explorer-backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/go-spacemesh/log"
)

type AccountUpdaterService interface {
	GetAccountState(address string) (uint64, uint64, error)
}

type Storage struct {
	NetworkInfo  model.NetworkInfo
	postUnitSize uint64

	client *mongo.Client
	db     *mongo.Database

	AccountUpdater AccountUpdaterService

	sync.Mutex
	changedEpoch int32
	lastEpoch    int32

	layersLock  sync.Mutex
	layersQueue *list.List
	layersReady *sync.Cond

	accountsLock  sync.Mutex
	accountsQueue map[uint32]map[string]bool
	accountsReady *sync.Cond
}

func New(parent context.Context, dbUrl string, dbName string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUrl))

	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		return nil, err
	}

	s := &Storage{
		client:        client,
		layersQueue:   list.New(),
		layersReady:   sync.NewCond(&sync.Mutex{}),
		accountsQueue: make(map[uint32]map[string]bool),
		accountsReady: sync.NewCond(&sync.Mutex{}),
		changedEpoch:  -1,
	}
	s.db = client.Database(dbName)

	err = s.InitAccountsStorage(ctx)
	if err != nil {
		log.Info("Init accounts storage error: %v", err)
	}
	err = s.InitAppsStorage(ctx)
	if err != nil {
		log.Info("Init apps storage error: %v", err)
	}
	err = s.InitActivationsStorage(ctx)
	if err != nil {
		log.Info("Init activations storage error: %v", err)
	}
	err = s.InitBlocksStorage(ctx)
	if err != nil {
		log.Info("Init blocks storage error: %v", err)
	}
	err = s.InitEpochsStorage(ctx)
	if err != nil {
		log.Info("Init epochs storage error: %v", err)
	}
	err = s.InitLayersStorage(ctx)
	if err != nil {
		log.Info("Init layers storage error: %v", err)
	}
	err = s.InitRewardsStorage(ctx)
	if err != nil {
		log.Info("Init rewards storage error: %v", err)
	}
	err = s.InitSmeshersStorage(ctx)
	if err != nil {
		log.Info("Init smeshers storage error: %v", err)
	}
	err = s.InitTransactionsStorage(ctx)
	if err != nil {
		log.Info("Init transactions storage error: %v", err)
	}

	go s.updateAccounts()
	go s.updateLayers()

	return s, nil
}

func (s *Storage) Close() {
	if s.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.db = nil
		err := s.client.Disconnect(ctx)
		if err != nil {
			log.Err(fmt.Errorf("error while disconnecting from database: %v", err))
		}
	}
}

func (s *Storage) OnNetworkInfo(genesisId string, genesisTime uint64, epochNumLayers uint32, maxTransactionsPerSecond uint64, layerDuration uint64, postUnitSize uint64) {
	s.NetworkInfo.GenesisId = genesisId
	s.NetworkInfo.GenesisTime = uint32(genesisTime)
	s.NetworkInfo.EpochNumLayers = epochNumLayers
	s.NetworkInfo.MaxTransactionsPerSecond = uint32(maxTransactionsPerSecond)
	s.NetworkInfo.LayerDuration = uint32(layerDuration)
	s.postUnitSize = postUnitSize

	err := s.SaveOrUpdateNetworkInfo(context.Background(), &s.NetworkInfo)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnNetworkInfo: error %v", err))
	}

	log.Info("Network Info: id: %s, genesis: %v, epoch layers: %v, max tx: %v, duration: %v",
		s.NetworkInfo.GenesisId,
		s.NetworkInfo.GenesisTime,
		s.NetworkInfo.EpochNumLayers,
		s.NetworkInfo.MaxTransactionsPerSecond,
		s.NetworkInfo.LayerDuration,
	)
}

func (s *Storage) OnNodeStatus(connectedPeers uint64, isSynced bool, syncedLayer uint32, topLayer uint32, verifiedLayer uint32) {
	s.NetworkInfo.ConnectedPeers = connectedPeers
	s.NetworkInfo.IsSynced = isSynced
	s.NetworkInfo.SyncedLayer = syncedLayer
	s.NetworkInfo.TopLayer = topLayer
	s.NetworkInfo.VerifiedLayer = verifiedLayer

	err := s.SaveOrUpdateNetworkInfo(context.Background(), &s.NetworkInfo)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnNodeStatus: error %v", err))
	}
}

func (s *Storage) GetEpochLayers(epoch int32) (uint32, uint32) {
	start := uint32(epoch) * s.NetworkInfo.EpochNumLayers
	end := start + s.NetworkInfo.EpochNumLayers - 1
	return start, end
}

func (s *Storage) GetEpochForLayer(layer uint32) uint32 {
	if s.NetworkInfo.EpochNumLayers > 0 {
		return layer / s.NetworkInfo.EpochNumLayers
	}
	return 0
}

func (s *Storage) OnLayer(in *pb.Layer) {
	s.pushLayer(in)
}

func (s *Storage) OnMalfeasanceProof(in *pb.MalfeasanceProof) {
	s.updateMalfeasanceProof(in)
}

func (s *Storage) OnAccount(in *pb.Account) {
	log.Info("OnAccount(%+v)", in)
	account := model.NewAccount(in)
	if account == nil {
		return
	}
	err := s.UpdateAccount(context.Background(), account.Address, account.Balance, account.Counter)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnAccount: error %v", err))
	}
}

func (s *Storage) OnReward(in *pb.Reward) {
	log.Info("OnReward(%+v)", in)
	reward := model.NewReward(in)
	if reward == nil {
		return
	}

	smesher, err := s.GetSmesher(context.Background(), &bson.D{{Key: "id", Value: in.Smesher}})
	// smesher, err := s.GetSmesherByCoinbase(context.Background(), reward.Coinbase)
	if err == nil {
		reward.Space = smesher.CommitmentSize
	}
	reward.Timestamp = s.getLayerTimestamp(reward.Layer)

	err = s.SaveReward(context.Background(), reward)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnReward save: error %v", err))
	}

	err = s.AddAccount(context.Background(), reward.Layer, reward.Coinbase, 0)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnReward add account: error %v", err))
	}

	//    s.AddAccountReward(context.Background(), reward.Layer, reward.Coinbase, reward.LayerReward, reward.Total - reward.LayerReward)
	err = s.AddAccountReward(context.Background(), reward.Layer, reward.Coinbase, reward.Total, reward.LayerReward)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnReward add account reward: error %v", err))
	}

	s.requestBalanceUpdate(reward.Layer, reward.Coinbase)
	s.setChangedEpoch(reward.Layer)
	s.updateEpochs() // trigger epoch stat recalculation todo: optimize this
}

func (s *Storage) OnTransactionReceipt(in *pb.TransactionReceipt) {
	log.Info("OnTransactionReceipt(%+v)", in)
	err := s.UpdateTransaction(context.Background(), model.NewTransactionReceipt(in))
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("OnTransactionReceipt: error %v", err))
	}
}

func (s *Storage) pushLayer(layer *pb.Layer) {
	s.layersLock.Lock()
	s.layersQueue.PushBack(layer)
	s.layersLock.Unlock()
	s.layersReady.Signal()
}

func (s *Storage) popLayer() *pb.Layer {
	s.layersLock.Lock()
	defer s.layersLock.Unlock()
	layer := s.layersQueue.Front()
	if layer != nil {
		return s.layersQueue.Remove(layer).(*pb.Layer)
	}
	return nil
}

func (s *Storage) requestBalanceUpdate(layer uint32, address string) {
	s.accountsLock.Lock()
	accounts, ok := s.accountsQueue[layer]
	if !ok {
		accounts = make(map[string]bool)
		s.accountsQueue[layer] = accounts
	}
	accounts[address] = true
	s.accountsLock.Unlock()
	s.accountsReady.Signal()
}

func (s *Storage) getAccountsQueue(accounts map[string]bool) int {
	s.accountsLock.Lock()
	defer s.accountsLock.Unlock()
	for layer, accs := range s.accountsQueue {
		if layer <= s.NetworkInfo.LastConfirmedLayer {
			for acc := range accs {
				accounts[acc] = true
			}
		}
		delete(s.accountsQueue, layer)
	}
	return len(accounts)
}

func (s *Storage) getChangedEpoch() int32 {
	s.Lock()
	defer s.Unlock()
	epoch := s.changedEpoch
	if s.changedEpoch >= 0 {
		s.changedEpoch = -1
	}
	return epoch
}

func (s *Storage) setChangedEpoch(layer uint32) {
	s.Lock()
	defer s.Unlock()
	if s.NetworkInfo.EpochNumLayers > 0 {
		epoch := int32(layer / s.NetworkInfo.EpochNumLayers)
		if s.changedEpoch < 0 || s.changedEpoch > epoch {
			s.changedEpoch = epoch
		}
		if epoch > s.lastEpoch {
			s.lastEpoch = epoch
		}
	}
}

func (s *Storage) updateLayer(in *pb.Layer) {
	layer, blocks, atxs, txs := model.NewLayer(in, &s.NetworkInfo)
	log.Info("updateLayer(%v) -> %v, %v, %v, %v, %v", in.Number.Number, layer.Number, len(blocks), len(atxs), len(txs), utils.BytesToHex(in.Hash))
	s.updateNetworkStatus(layer)

	err := s.SaveOrUpdateBlocks(context.Background(), blocks)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("updateLayer: error %v", err))
	}

	s.updateActivations(layer, atxs)
	s.updateTransactions(layer, txs)

	err = s.SaveOrUpdateLayer(context.Background(), layer)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("updateLayer: error %v", err))
	}

	s.setChangedEpoch(layer.Number)
	s.accountsReady.Signal()
	s.updateEpochs()
}

func (s *Storage) updateNetworkStatus(layer *model.Layer) {
	s.NetworkInfo.LastLayer = layer.Number
	s.NetworkInfo.LastLayerTimestamp = uint32(time.Now().Unix())
	if layer.Status == int(pb.Layer_LAYER_STATUS_APPROVED) {
		s.NetworkInfo.LastApprovedLayer = layer.Number
	} else if layer.Status == int(pb.Layer_LAYER_STATUS_CONFIRMED) {
		s.NetworkInfo.LastConfirmedLayer = layer.Number
	}

	err := s.SaveOrUpdateNetworkInfo(context.Background(), &s.NetworkInfo)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("updateNetworkStatus: error %v", err))
	}
}

func (s *Storage) updateActivations(layer *model.Layer, atxs []*model.Activation) {
	log.Info("updateActivations(%v)", len(atxs))
	err := s.SaveOrUpdateActivations(context.Background(), atxs)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("updateActivations: error %v", err))
	}

	var coinbaseUpdateOps []mongo.WriteModel
	var smesherUpdateOps []mongo.WriteModel

	for _, atx := range atxs {
		//err := s.SaveSmesher(context.Background(), atx.GetSmesher(s.postUnitSize))
		////TODO: better error handling
		//if err != nil {
		//	log.Err(fmt.Errorf("updateActivations: error %v", err))
		//}
		smesherUpdateOps = append(smesherUpdateOps, s.SaveSmesherQuery(atx.GetSmesher(s.postUnitSize)))

		//err = s.UpdateSmesher(context.Background(), atx.SmesherId, atx.Coinbase, uint64(atx.NumUnits)*s.postUnitSize, s.getLayerTimestamp(atx.Layer))
		////TODO: better error handling
		//if err != nil {
		//	log.Err(fmt.Errorf("updateActivations: error %v", err))
		//}
		coinbaseOp, smesherOp := s.UpdateSmesherQuery(atx.SmesherId, atx.Coinbase, uint64(atx.NumUnits)*s.postUnitSize, s.getLayerTimestamp(atx.Layer))
		coinbaseUpdateOps = append(coinbaseUpdateOps, coinbaseOp)
		smesherUpdateOps = append(smesherUpdateOps, smesherOp)

		err = s.AddAccount(context.Background(), layer.Number, atx.Coinbase, 0)
		//TODO: better error handling
		if err != nil {
			log.Err(fmt.Errorf("updateActivations: error %v", err))
		}
	}

	if len(smesherUpdateOps) > 0 {
		_, err = s.db.Collection("smeshers").BulkWrite(context.TODO(), smesherUpdateOps)
		if err != nil {
			log.Err(fmt.Errorf("updateActivations: error smeshers write %v", err))
		}
	}

	if len(coinbaseUpdateOps) > 0 {
		_, err = s.db.Collection("coinbases").BulkWrite(context.TODO(), coinbaseUpdateOps)
		if err != nil {
			log.Err(fmt.Errorf("updateActivations: error smeshers write %v", err))
		}
	}
}

func (s *Storage) updateTransactions(layer *model.Layer, txs map[string]*model.Transaction) {
	log.Info("updateTransactions")
	for _, tx := range txs {
		err := s.SaveTransaction(context.Background(), tx)
		if err != nil {
			continue
		}

		if tx.Sender != "" {
			err := s.AddAccount(context.Background(), layer.Number, tx.Sender, 0)
			//TODO: better error handling
			if err != nil {
				log.Err(fmt.Errorf("updateTransactions: error %v", err))
			}

			err = s.AddAccountSent(context.Background(), layer.Number, tx.Sender, tx.Amount, tx.Fee)
			//TODO: better error handling
			if err != nil {
				log.Err(fmt.Errorf("updateTransactions: error %v", err))
			}

			s.requestBalanceUpdate(layer.Number, tx.Sender)
		}
		if tx.Receiver != "" {
			err := s.AddAccount(context.Background(), layer.Number, tx.Receiver, 0)
			//TODO: better error handling
			if err != nil {
				log.Err(fmt.Errorf("updateTransactions: error %v", err))
			}

			err = s.AddAccountReceived(context.Background(), layer.Number, tx.Receiver, tx.Amount)
			//TODO: better error handling
			if err != nil {
				log.Err(fmt.Errorf("updateTransactions: error %v", err))
			}
			s.requestBalanceUpdate(layer.Number, tx.Receiver)
		}
	}
}

func (s *Storage) updateEpoch(epochNumber int32, prev *model.Epoch) *model.Epoch {
	log.Info("updateEpoch(%v)", epochNumber)
	epoch := &model.Epoch{Number: epochNumber}
	s.computeStatistics(epoch)
	if prev != nil {
		epoch.Stats.Cumulative.Capacity = epoch.Stats.Current.Capacity
		epoch.Stats.Cumulative.Decentral = prev.Stats.Current.Decentral
		//        epoch.Stats.Cumulative.Smeshers      = prev.Stats.Current.Smeshers
		epoch.Stats.Cumulative.Smeshers = epoch.Stats.Current.Smeshers
		epoch.Stats.Cumulative.Transactions = prev.Stats.Cumulative.Transactions + epoch.Stats.Current.Transactions
		epoch.Stats.Cumulative.Accounts = epoch.Stats.Current.Accounts
		epoch.Stats.Cumulative.Rewards = prev.Stats.Cumulative.Rewards + epoch.Stats.Current.Rewards
		epoch.Stats.Cumulative.RewardsNumber = prev.Stats.Cumulative.RewardsNumber + epoch.Stats.Current.RewardsNumber
		epoch.Stats.Cumulative.Security = prev.Stats.Current.Security
		epoch.Stats.Cumulative.TxsAmount = prev.Stats.Cumulative.TxsAmount + epoch.Stats.Current.TxsAmount
		epoch.Stats.Current.Circulation = epoch.Stats.Cumulative.Rewards
		epoch.Stats.Cumulative.Circulation = epoch.Stats.Current.Circulation
	} else {
		epoch.Stats.Current.Circulation = epoch.Stats.Current.Rewards
		epoch.Stats.Cumulative = epoch.Stats.Current
	}
	err := s.SaveOrUpdateEpoch(context.Background(), epoch)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("updateEpoch: error %v", err))
	}

	return epoch
}

func (s *Storage) updateEpochs() {
	epochNumber := s.getChangedEpoch()
	if epochNumber >= 0 {
		var prev *model.Epoch
		if epochNumber > 0 {
			prev, _ = s.GetEpochByNumber(context.Background(), epochNumber-1)
		}
		for i := epochNumber; i <= s.lastEpoch; i++ {
			prev = s.updateEpoch(i, prev)
		}
	}
}

func (s *Storage) updateAccount(address string) {
	balance, counter, err := s.AccountUpdater.GetAccountState(address)
	if err != nil {
		return
	}
	log.Info("Update account %v: balance %v, counter %v", address, balance, counter)

	err = s.UpdateAccount(context.Background(), address, balance, counter)
	//TODO: better error handling
	if err != nil {
		log.Err(fmt.Errorf("updateEpoch: error %v", err))
	}
}

func (s *Storage) updateLayers() {
	for {
		s.layersReady.L.Lock()
		s.layersReady.Wait()
		s.layersReady.L.Unlock()

		for layer := s.popLayer(); layer != nil; layer = s.popLayer() {
			s.updateLayer(layer)
		}
	}
}

func (s *Storage) updateAccounts() {
	for {
		s.accountsReady.L.Lock()
		s.accountsReady.Wait()
		s.accountsReady.L.Unlock()

		accounts := make(map[string]bool)
		if s.getAccountsQueue(accounts) > 0 {
			for address := range accounts {
				s.updateAccount(address)
			}
		}
	}
}

func (s *Storage) updateMalfeasanceProof(in *pb.MalfeasanceProof) {
	proof := model.NewMalfeasanceProof(in)
	if proof == nil {
		return
	}

	log.Info("updateMalfeasanceProof -> %v, %v, %v", proof.Layer, proof.Smesher, proof.Kind)

	err := s.SaveMalfeasanceProof(context.Background(), proof)
	if err != nil {
		log.Err(fmt.Errorf("updateMalfeasanceProof: %v", err))
	}
}

func (s *Storage) GetEpochLayersFilter(epochNumber int32, key string) *bson.D {
	layerStart, layerEnd := s.GetEpochLayers(epochNumber)
	return &bson.D{{Key: key, Value: bson.D{{Key: "$gte", Value: layerStart}, {Key: "$lte", Value: layerEnd}}}}
}

func (s *Storage) getLayerTimestamp(layer uint32) uint32 {
	if layer == 0 {
		return s.NetworkInfo.GenesisTime
	}
	return s.NetworkInfo.GenesisTime + layer*s.NetworkInfo.LayerDuration
}

func (s *Storage) Ping() error {
	if s.client == nil {
		return errors.New("Storage not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return s.client.Ping(ctx, nil)
}
