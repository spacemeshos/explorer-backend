package storage

import (
    "container/list"
    "context"
    "errors"
    "time"
    "sync"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/model"
)

type AccountUpdaterService interface {
    GetAccountState(address string) (uint64, uint64, error)
}

type Storage struct {
    NetworkInfo		model.NetworkInfo

    client		*mongo.Client
    db			*mongo.Database

    AccountUpdater	AccountUpdaterService

    sync.Mutex
    changedEpoch	int32
    lastEpoch		int32

    layersLock		sync.Mutex
    layersQueue		*list.List
    layersReady		*sync.Cond

    accountsLock	sync.Mutex
    accountsQueue	map[uint32]map[string]bool
    accountsReady	*sync.Cond
}

func New(parent context.Context, dbUrl string, dbName string) (*Storage, error) {
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUrl))

    if err != nil {
        return nil, err
    }

    err = client.Ping(ctx, nil)

    if err != nil {
        return nil, err
    }

    s := &Storage{
        client: client,
        layersQueue: list.New(),
        layersReady: sync.NewCond(&sync.Mutex{}),
        accountsQueue: make(map[uint32]map[string]bool),
        accountsReady: sync.NewCond(&sync.Mutex{}),
        changedEpoch: -1,
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
        ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
        s.db = nil
        s.client.Disconnect(ctx)
    }
}

func (s *Storage) OnNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64) {
    s.NetworkInfo.NetId = uint32(netId)
    s.NetworkInfo.GenesisTime = uint32(genesisTime)
    s.NetworkInfo.EpochNumLayers = uint32(epochNumLayers)
    s.NetworkInfo.MaxTransactionsPerSecond = uint32(maxTransactionsPerSecond)
    s.NetworkInfo.LayerDuration = uint32(layerDuration)

    s.SaveOrUpdateNetworkInfo(context.Background(), &s.NetworkInfo)

    log.Info("Network Info: id: %v, genesis: %v, epoch layers: %v, max tx: %v, duration: %v",
        s.NetworkInfo.NetId,
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

    s.SaveOrUpdateNetworkInfo(context.Background(), &s.NetworkInfo)
}

func (s *Storage) GetEpochLayers(epoch int32) (uint32, uint32) {
    start := uint32(epoch) * uint32(s.NetworkInfo.EpochNumLayers)
    end := start + uint32(s.NetworkInfo.EpochNumLayers) - 1
    return start, end
}

func (s *Storage) GetEpochForLayer(layer uint32) uint32 {
    if s.NetworkInfo.EpochNumLayers > 0 {
        return layer / uint32(s.NetworkInfo.EpochNumLayers)
    }
    return 0
}

func (s *Storage) OnLayer(in *pb.Layer) {
    s.pushLayer(in)
}

func (s *Storage) OnAccount(in *pb.Account) {
    log.Info("OnAccount(%+v)", in)
    account := model.NewAccount(in)
    if account == nil {
        return
    }
    s.UpdateAccount(context.Background(), account.Address, account.Balance, account.Counter)
}

func (s *Storage) OnReward(in *pb.Reward) {
    log.Info("OnReward(%+v)", in)
    reward := model.NewReward(in)
    if reward == nil {
        return
    }
    smesher, err := s.GetSmesherByCoinbase(context.Background(), reward.Coinbase)
    if err == nil {
        reward.Smesher = smesher.Id
        reward.Space = smesher.CommitmentSize
    }
    reward.Timestamp = s.getLayerTimestamp(reward.Layer)
    s.SaveReward(context.Background(), reward)
    s.AddAccount(context.Background(), reward.Layer, reward.Coinbase, 0)
//    s.AddAccountReward(context.Background(), reward.Layer, reward.Coinbase, reward.LayerReward, reward.Total - reward.LayerReward)
    s.AddAccountReward(context.Background(), reward.Layer, reward.Coinbase, reward.Total, reward.LayerReward)
    s.requestBalanceUpdate(reward.Layer, reward.Coinbase)
    s.setChangedEpoch(reward.Layer)
}

func (s *Storage) OnTransactionReceipt(in *pb.TransactionReceipt) {
    log.Info("OnTransactionReceipt(%+v)", in)
    s.UpdateTransaction(context.Background(), model.NewTransactionReceipt(in))
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
            for acc, _ := range accs {
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
    log.Info("updateLayer(%v) -> %v, %v, %v, %v", in.Number.Number, layer.Number, len(blocks), len(atxs), len(txs))
    s.updateNetworkStatus(layer)
    s.SaveOrUpdateBlocks(context.Background(), blocks)
    s.updateActivations(layer, atxs)
    s.updateTransactions(layer, txs)
    s.updateLayerRewards(layer)
    s.SaveOrUpdateLayer(context.Background(), layer)
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
    s.SaveOrUpdateNetworkInfo(context.Background(), &s.NetworkInfo)
}

func (s *Storage) updateActivations(layer *model.Layer, atxs []*model.Activation) {
    log.Info("updateActivations(%v)", len(atxs))
    s.SaveOrUpdateActivations(context.Background(), atxs)
    for _, atx := range atxs {
        s.SaveSmesher(context.Background(), atx.GetSmesher())
        s.UpdateSmesher(context.Background(), atx.SmesherId, atx.Coinbase, atx.CommitmentSize, s.getLayerTimestamp(atx.Layer))
        s.AddAccount(context.Background(), layer.Number, atx.Coinbase, 0)
    }
}

func (s *Storage) updateTransactions(layer *model.Layer, txs map[string]*model.Transaction) {
    log.Info("updateTransactions")
    s.SaveTransactions(context.Background(), txs)
    for _, tx := range txs {
        if  tx.Sender != "" {
            s.AddAccount(context.Background(), layer.Number, tx.Sender, 0)
            s.AddAccountSent(context.Background(), layer.Number, tx.Sender, tx.Amount)
            s.requestBalanceUpdate(layer.Number, tx.Sender)
        }
        if tx.Receiver != "" {
            s.AddAccount(context.Background(), layer.Number, tx.Receiver, 0)
            s.AddAccountReceived(context.Background(), layer.Number, tx.Receiver, tx.Amount)
            s.requestBalanceUpdate(layer.Number, tx.Receiver)
        }
    }
}

func (s *Storage) updateLayerRewards(layer *model.Layer) {
    log.Info("updateLayerRewards")
    rewards, _ := s.GetLayersRewards(context.Background(), layer.Number, layer.Number)
    layer.Rewards = uint64(rewards)
}

func (s *Storage) updateEpoch(epochNumber int32, prev *model.Epoch) *model.Epoch {
    log.Info("updateEpoch(%v)", epochNumber)
    epoch := &model.Epoch{Number: epochNumber}
    s.computeStatistics(epoch)
    if prev != nil {
        epoch.Stats.Cumulative.Capacity      = epoch.Stats.Current.Capacity
        epoch.Stats.Cumulative.Decentral     = prev.Stats.Current.Decentral
//        epoch.Stats.Cumulative.Smeshers      = prev.Stats.Current.Smeshers
        epoch.Stats.Cumulative.Smeshers      = epoch.Stats.Current.Smeshers
        epoch.Stats.Cumulative.Transactions  = prev.Stats.Cumulative.Transactions + epoch.Stats.Current.Transactions
        epoch.Stats.Cumulative.Accounts      = epoch.Stats.Current.Accounts
        epoch.Stats.Cumulative.Rewards       = prev.Stats.Cumulative.Rewards + epoch.Stats.Current.Rewards
        epoch.Stats.Cumulative.RewardsNumber = prev.Stats.Cumulative.RewardsNumber + epoch.Stats.Current.RewardsNumber
        epoch.Stats.Cumulative.Security      = prev.Stats.Current.Security
        epoch.Stats.Cumulative.TxsAmount     = prev.Stats.Cumulative.TxsAmount + epoch.Stats.Current.TxsAmount
        epoch.Stats.Current.Circulation      = epoch.Stats.Cumulative.Rewards
        epoch.Stats.Cumulative.Circulation   = epoch.Stats.Current.Circulation
    } else {
        epoch.Stats.Current.Circulation = epoch.Stats.Current.Rewards
        epoch.Stats.Cumulative = epoch.Stats.Current
    }
    s.SaveOrUpdateEpoch(context.Background(), epoch)
    return epoch
}

func (s *Storage) updateEpochs() {
    epochNumber := s.getChangedEpoch()
    if epochNumber >= 0 {
        var prev *model.Epoch
        if epochNumber > 0 {
            prev, _ = s.GetEpochByNumber(context.Background(), epochNumber - 1)
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
    s.UpdateAccount(context.Background(), address, balance, counter)
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
            for address, _ := range accounts {
                s.updateAccount(address)
            }
        }
    }
}

func (s *Storage) GetEpochLayersFilter(epochNumber int32, key string) *bson.D {
    layerStart, layerEnd := s.GetEpochLayers(epochNumber)
    return &bson.D{{key, bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}
}

func (s *Storage) getLayerTimestamp(layer uint32) uint32 {
    if layer == 0 {
        return s.NetworkInfo.GenesisTime
    }
    return s.NetworkInfo.GenesisTime + (layer - 1) * s.NetworkInfo.LayerDuration
}

func (s *Storage) Ping() error {
    if s.client == nil {
        return errors.New("Storage not initialized")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    return s.client.Ping(ctx, nil)
}
