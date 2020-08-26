package storage

import (
    "context"
    "time"

//    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/model"
)

type Storage struct {
    NetId                    uint64
    GenesisTime              uint64
    EpochNumLayers           uint64
    MaxTransactionsPerSecond uint64
    LayerDuration            uint64

    client *mongo.Client
    db *mongo.Database
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
    s.EpochNumLayers = epochNumLayers
    s.MaxTransactionsPerSecond = maxTransactionsPerSecond
    s.LayerDuration = layerDuration
}

func (s *Storage) GetEpochLayers(epoch uint32) (uint32, uint32) {
    start := epoch * uint32(s.EpochNumLayers)
    end := start + uint32(s.EpochNumLayers) - 1
    return start, end
}

func (s *Storage) OnLayer(in *pb.Layer) {
    if model.IsConfirmedLayer(in) {
        layer, blocks, atxs, txs := model.NewLayer(in, s.GenesisTime, s.LayerDuration)
        s.SaveLayer(context.Background(), layer)
        s.SaveBlocks(context.Background(), blocks)
        s.SaveActivations(context.Background(), atxs)
        s.SaveTransactions(context.Background(), txs)
    }
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
