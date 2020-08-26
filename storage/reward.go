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

func (s *Storage) InitRewardsStorage(ctx context.Context) error {
    models := []mongo.IndexModel{
        {Keys: bson.D{{"layer", 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)},
        {Keys: bson.D{{"smesher", 1}}, Options: options.Index().SetName("smesherIndex").SetUnique(false)},
        {Keys: bson.D{{"coinbase", 1}}, Options: options.Index().SetName("coinbaseIndex").SetUnique(false)},
    }
    _, err := s.db.Collection("rewards").Indexes().CreateMany(ctx, models, options.CreateIndexes().SetMaxTime(2 * time.Second));
    return err
}

func (s *Storage) GetReward(parent context.Context, query *bson.D) (*model.Reward, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("rewards").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Reward{
        Layer: uint32(doc.Lookup("layer").Int32()),
        Total: uint64(doc.Lookup("total").Int64()),
        LayerReward: uint64(doc.Lookup("layerReward").Int64()),
        LayerComputed: uint32(doc.Lookup("layerComputed").Int32()),
        Coinbase: doc.Lookup("coinbase").String(),
        Smesher: doc.Lookup("smesher").String(),
    }
    return account, nil
}

func (s *Storage) GetRewardsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    return s.db.Collection("rewards").CountDocuments(ctx, query, opts...)
}

func (s *Storage) GetRewards(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("rewards").Find(ctx, query, opts...)
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

func (s *Storage) SaveReward(parent context.Context, in *model.Reward) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("rewards").InsertOne(ctx, bson.D{
        {"layer", in.Layer},
        {"total", in.Total},
        {"layerReward", in.LayerReward},
        {"layerComputed", in.LayerComputed},
        {"coinbase", in.Coinbase},
        {"smesher", in.Smesher},
    })
    return err
}
