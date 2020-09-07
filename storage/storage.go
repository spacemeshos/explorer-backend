package storage

import (
    "container/list"
    "context"
    "time"
    "sync"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/model"
)

type Storage struct {
    NetId                    uint64
    GenesisTime              uint64
    EpochNumLayers           uint32
    MaxTransactionsPerSecond uint64
    LayerDuration            uint64

    lastLayer		uint32
    lastEpoch		int32

    client	*mongo.Client
    db		*mongo.Database

    sync.Mutex
    queue	*list.List
    ready	*sync.Cond
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
        queue: list.New(),
        ready: sync.NewCond(&sync.Mutex{}),
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

    go s.update()

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
    s.NetId = netId
    s.GenesisTime = genesisTime
    s.EpochNumLayers = uint32(epochNumLayers)
    s.MaxTransactionsPerSecond = maxTransactionsPerSecond
    s.LayerDuration = layerDuration
}

func (s *Storage) GetEpochLayers(epoch int32) (uint32, uint32) {
    start := uint32(epoch) * uint32(s.EpochNumLayers)
    end := start + uint32(s.EpochNumLayers) - 1
    return start, end
}

func (s *Storage) GetEpochForLayer(layer uint32) uint32 {
    if s.EpochNumLayers > 0 {
        return layer / uint32(s.EpochNumLayers)
    }
    return 0
}

func (s *Storage) OnLayer(in *pb.Layer) {
    if model.IsConfirmedLayer(in) {
        layer, blocks, atxs, txs := model.NewLayer(in, s.GenesisTime, s.LayerDuration)
        s.SaveLayer(context.Background(), layer)
        s.SaveBlocks(context.Background(), blocks)
        s.SaveActivations(context.Background(), atxs)
        s.SaveTransactions(context.Background(), txs)
    }
    s.pushLayer(in)
}

func (s *Storage) OnAccount(in *pb.Account) {
    account := model.NewAccount(in)
    if account == nil {
        return
    }
    s.SaveAccount(context.Background(), account)
}

func (s *Storage) OnReward(in *pb.Reward) {
    reward := model.NewReward(in)
    if reward == nil {
        return
    }
    s.SaveReward(context.Background(), reward)
}

func (s *Storage) OnTransactionReceipt(in *pb.TransactionReceipt) {
    s.UpdateTransaction(context.Background(), model.NewTransactionReceipt(in))
}

func (s *Storage) pushLayer(layer *pb.Layer) {
    s.Lock()
    s.queue.PushBack(layer)
    s.Unlock()
    s.ready.Signal()
}

func (s *Storage) popLayer() *pb.Layer {
    s.Lock()
    defer s.Unlock()
    layer := s.queue.Front()
    if layer != nil {
        return s.queue.Remove(layer).(*pb.Layer)
    }
    return nil
}

func (s *Storage) updateLayer(in *pb.Layer) {
    log.Info("updateLayer(%v)", in.Number.Number)
    layer, blocks, atxs, txs := model.NewLayer(in, s.GenesisTime, s.LayerDuration)
    s.SaveOrUpdateBlocks(context.Background(), blocks)
    s.updateActivations(atxs)
    s.SaveTransactions(context.Background(), txs)
    s.updateLayerRewards(layer)
    s.SaveOrUpdateLayer(context.Background(), layer)
    s.updateEpochsFrom(int32(layer.Number / s.EpochNumLayers))
}

func (s *Storage) updateActivations(atxs []*model.Activation) {
    s.SaveOrUpdateActivations(context.Background(), atxs)
    for _, atx := range atxs {
        s.SaveOrUpdateSmesher(context.Background(), atx.GetSmesher())
    }
}

func (s *Storage) updateLayerRewards(layer *model.Layer) {
    layer.Rewards = uint64(s.GetLayersRewards(context.Background(), layer.Number, layer.Number))
}

func (s *Storage) updateEpoch(epochNumber int32, prev *model.Epoch) *model.Epoch {
    epoch := &model.Epoch{Number: epochNumber}
    s.computeStatistics(epoch)
    if prev != nil {
        epoch.Stats.Cumulative.Capacity     = epoch.Stats.Current.Capacity
        epoch.Stats.Cumulative.Decentral    = epoch.Stats.Current.Decentral
        epoch.Stats.Cumulative.Smeshers     = epoch.Stats.Current.Smeshers
        epoch.Stats.Cumulative.Transactions = prev.Stats.Cumulative.Transactions + epoch.Stats.Current.Transactions
        epoch.Stats.Cumulative.Accounts     = epoch.Stats.Current.Accounts
        epoch.Stats.Cumulative.Circulation  = epoch.Stats.Current.Circulation
        epoch.Stats.Cumulative.Rewards      = prev.Stats.Cumulative.Rewards + epoch.Stats.Current.Rewards
        epoch.Stats.Cumulative.Security     = prev.Stats.Current.Security
    } else {
        epoch.Stats.Cumulative = epoch.Stats.Current
    }
    s.SaveOrUpdateEpoch(context.Background(), epoch)
    return epoch
}

func (s *Storage) updateEpochsFrom(epochNumber int32) {
    if epochNumber > s.lastEpoch {
        s.lastEpoch = epochNumber
    }
    var prev *model.Epoch
    if epochNumber > 0 {
        prev, _ = s.GetEpochByNumber(context.Background(), epochNumber - 1)
    }
    for i := epochNumber; i <= s.lastEpoch; i++ {
        prev = s.updateEpoch(i, prev)
    }
}

func (s *Storage) update() {
    for {
        s.ready.L.Lock()
        s.ready.Wait()
        s.ready.L.Unlock()

        layer := s.popLayer()
        if layer != nil {
            s.updateLayer(layer)
        }
    }
}

func (s *Storage) GetEpochLayersFilter(epochNumber int32, key string) *bson.D {
    layerStart, layerEnd := s.GetEpochLayers(epochNumber)
    return &bson.D{{key, bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}}}
}
