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
        Address: doc.Lookup("address").String(),
        Balance: utils.GetAsUInt64(doc.Lookup("balance")),
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

func (s *Storage) SaveAccount(parent context.Context, in *model.Account) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("accounts").InsertOne(ctx, bson.D{
        {"address", in.Address},
        {"balance", in.Balance},
    })
    if err != nil {
        log.Info("SaveAccounts: %v", err)
    }
    return err
}

func (s *Storage) SaveOrUpdateAccount(parent context.Context, in *model.Account) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{"address", in.Address}}, bson.D{
        {"$set", bson.D{
            {"address", in.Address},
            {"balance", in.Balance},
        }},
    }, options.Update().SetUpsert(true))
    if err != nil {
        log.Info("SaveOrUpdateAccounts: %v", err)
    }
    return err
}
