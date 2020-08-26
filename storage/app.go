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

func (s *Storage) InitAppsStorage(ctx context.Context) error {
    _, err := s.db.Collection("apps").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"address", 1}}, Options: options.Index().SetName("addressIndex").SetUnique(true)});
    return err
}

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

func (s *Storage) GetApps(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.App, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("apps").Find(ctx, query, opts...)
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
    accounts := make([]*model.App, len(docs.([]bson.D)), len(docs.([]bson.D)))
    for i, doc := range docs.([]bson.D) {
        accounts[i] = &model.App{
            Address: doc[0].Value.(string),
        }
    }
    return accounts, nil
}

func (s *Storage) SaveApp(parent context.Context, in *model.App) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("apps").InsertOne(ctx, bson.D{
        {"address", in.Address},
    })
    return err
}
