package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/explorer-backend/model"
    "github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) InitAccountsStorage(ctx context.Context) error {
    _, err := s.db.Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"address", 1}}, Options: options.Index().SetName("addressIndex").SetUnique(true)});
    _, err = s.db.Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"layer", 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)});
    _, err = s.db.Collection("ledger").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"address", 1}}, Options: options.Index().SetName("addressIndex").SetUnique(false)});
    _, err = s.db.Collection("ledger").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"layer", 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)});
    return err
}

func (s *Storage) GetAccount(parent context.Context, query *bson.D) (*model.Account, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("accounts").Find(ctx, query)
    if err != nil {
        log.Info("GetAccount: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetAccount: Empty result", err)
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Account{
        Address: utils.GetAsString(doc.Lookup("address")),
        Balance: utils.GetAsUInt64(doc.Lookup("balance")),
        Counter: utils.GetAsUInt64(doc.Lookup("counter")),
    }
    return account, nil
}

func (s *Storage) GetAccountsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("accounts").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetAccountsCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) GetAccounts(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("accounts").Find(ctx, query, opts...)
    if err != nil {
        log.Info("GetAccounts: %v", err)
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        log.Info("GetAccounts: %v", err)
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        log.Info("GetAccounts: Empty result")
        return nil, nil
    }
    return docs.([]bson.D), nil
}

func (s *Storage) AddAccount(parent context.Context, layer uint32, address string, balance uint64) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("accounts").InsertOne(ctx, bson.D{
        {"address", address},
        {"layer", layer},
        {"balance", balance},
        {"counter", uint64(0)},
    })
    if err != nil {
        log.Info("AddAccount: %v", err)
    }
    return nil
}

func (s *Storage) SaveAccount(parent context.Context, layer uint32, in *model.Account) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("accounts").InsertOne(ctx, bson.D{
        {"address", in.Address},
        {"layer", layer},
        {"balance", in.Balance},
        {"counter", in.Counter},
    })
    if err != nil {
        log.Info("SaveAccount: %v", err)
    }
    return nil
}

func (s *Storage) UpdateAccount(parent context.Context, address string, balance uint64, counter uint64) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{"address", address}}, bson.D{
        {"$set", bson.D{
            {"balance", balance},
            {"counter", counter},
        }},
    })
    if err != nil {
        log.Info("UpdateAccount: %v", err)
    }
    return nil
}

func (s *Storage) AddAccountSent(parent context.Context, layer uint32, address string, amount uint64) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("ledger").InsertOne(ctx, bson.D{
        {"address", address},
        {"layer", layer},
        {"sent", amount},
    })
    if err != nil {
        log.Info("AddAccountSent: %v", err)
    }
    return nil
}

func (s *Storage) AddAccountReceived(parent context.Context, layer uint32, address string, amount uint64) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("ledger").InsertOne(ctx, bson.D{
        {"address", address},
        {"layer", layer},
        {"received", amount},
    })
    if err != nil {
        log.Info("AddAccountReceived: %v", err)
    }
    return nil
}

func (s *Storage) AddAccountReward(parent context.Context, layer uint32, address string, reward uint64, fee uint64) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    status, err := s.db.Collection("ledger").InsertOne(ctx, bson.D{
        {"address", address},
        {"layer", layer},
        {"reward", reward},
        {"fee", fee},
    })
    if err != nil {
        log.Info("AddAccountReward: %v", err)
    } else {
        log.Info("AddAccountReward: %v", status)
    }
    return nil
}

func (s *Storage) GetAccountSummary(parent context.Context, address string) (uint64, uint64, uint64, uint64, uint32) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    matchStage := bson.D{
        {"$match", bson.D{
            {"address", address},
        }},
    }
    groupStage := bson.D{
        {"$group", bson.D{
            {"_id", ""},
            {"sent", bson.D{
                {"$sum", "$sent"},
            }},
            {"reveived", bson.D{
                {"$sum", "$reveived"},
            }},
            {"awards", bson.D{
                {"$sum", "$reward"},
            }},
            {"fees", bson.D{
                {"$sum", "$fee"},
            }},
            {"layer", bson.D{
                {"$max", "$layer"},
            }},
        }},
    }
    cursor, err := s.db.Collection("ledger").Aggregate(ctx, mongo.Pipeline{
        matchStage,
        groupStage,
    })
    if err != nil {
        log.Info("GetAccountSummary: %v", err)
        return 0, 0, 0, 0, 0
    }
    if !cursor.Next(ctx) {
        log.Info("GetAccountSummary: Empty result")
        return 0, 0, 0, 0, 0
    }
    doc := cursor.Current
    return utils.GetAsUInt64(doc.Lookup("sent")), utils.GetAsUInt64(doc.Lookup("received")), utils.GetAsUInt64(doc.Lookup("awards")), utils.GetAsUInt64(doc.Lookup("fees")), s.NetworkInfo.GenesisTime + s.NetworkInfo.LayerDuration * utils.GetAsUInt32(doc.Lookup("layer"))
}
