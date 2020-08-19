package model

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/explorer-backend/model"
)

func (s *Storage) GetAccount(parent context.Context, query *bson.D) (*Account, error) {
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

func (s *Storage) GetAccounts(parent context.Context, query *bson.D) ([]*Account, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("accounts").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        return nil, err
    }
    if len(docs) == 0 {
        return nil, nil
    }
    accounts := make([]*model.Account, len(docs), len(docs))
    for i, doc := range docs {
        accounts[i] = &model.Account{
            Address: doc.Lookup("address").String(),
            Balance: uint64(doc.Lookup("balance").Int64()),
        }
    }
    return accounts, nil
}

func (s *Storage) SaveAccount(parent context.Context, in *Account) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("accounts").InsertOne(ctx, bson.D{
        {"address", in.Address},
        {"balance", in.Balance},
    })
    return err
}
