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
)

func (s *Storage) GetApp(parent context.Context, query *bson.D) (*model.App, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("apps").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.App{
        Address: doc.Lookup("address").String(),
    }
    return account, nil
}

func (s *Storage) GetApps(parent context.Context, query *bson.D) ([]*model.App, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("apps").Find(ctx, query)
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
    accounts := make([]*model.App, len(docs), len(docs))
    for i, doc := range docs {
        accounts[i] = &model.App{
            Address: doc.Lookup("address").String(),
        }
    }
    return accounts, nil
}

func (s *Storage) SaveApp(parent context.Context, in *model.App) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("apps").InsertOne(ctx, bson.D{
        {"address", in.Address},
    })
    return err
}
