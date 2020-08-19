package storage

import (
    "context"
    "time"

//    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
//    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/model"
)

type Storage struct {
    client *mongo.Client
    db *mongo.Database
}

func New() *Storage {
    return &Storage{}
}

func (s *Storage) Open(dbUrl string, dbName string) error {
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUrl))

    if err != nil {
        return err
    }

    err = client.Ping(ctx, nil)

    if err != nil {
        return err
    }

    s.client = client
    s.db = client.Database(dbName)

    return nil
}

func (s *Storage) Close() {
    if s.client != nil {
        ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
        s.db = nil
        s.client.Disconnect(ctx)
    }
}

func (s *Storage) OnNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64) {
}

func (s *Storage) OnLayer(in *pb.Layer) {
    if model.IsConfirmedLayer(in) {
        layer, blocks, atxs, txs := model.NewLayer(in)
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
