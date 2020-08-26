package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/explorer-backend/model"
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
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Account{
        Address: doc.Lookup("address").String(),
        Balance: uint64(doc.Lookup("balance").Int64()),
    }
    return account, nil
}

func (s *Storage) GetAccountsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    return s.db.Collection("accounts").CountDocuments(ctx, query, opts...)
}

func (s *Storage) GetAccounts(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("accounts").Find(ctx, query, opts...)
    if err != nil {
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
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
    return err
}
