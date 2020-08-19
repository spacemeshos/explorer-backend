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
        Layer: uint64(doc.Lookup("layer").Int64()),
        Total: uint64(doc.Lookup("total").Int64()),
        LayerReward: uint64(doc.Lookup("layerReward").Int64()),
        LayerComputed: uint64(doc.Lookup("layerComputed").Int64()),
        Coinbase: doc.Lookup("coinbase").String(),
        Smesher: doc.Lookup("smesher").String(),
    }
    return account, nil
}

func (s *Storage) GetRewards(parent context.Context, query *bson.D) ([]*model.Reward, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("rewards").Find(ctx, query)
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
    rewards := make([]*model.Reward, len(docs), len(docs))
    for i, doc := range docs {
        rewards[i] = &model.Reward{
            Layer: uint64(doc.Lookup("layer").Int64()),
            Total: uint64(doc.Lookup("total").Int64()),
            LayerReward: uint64(doc.Lookup("layerReward").Int64()),
            LayerComputed: uint64(doc.Lookup("layerComputed").Int64()),
            Coinbase: doc.Lookup("coinbase").String(),
            Smesher: doc.Lookup("smesher").String(),
        }
    }
    return rewards, nil
}

func (s *Storage) SaveReward(parent context.Context, in *model.Reward) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("rewards").InsertOne(ctx, bson.D{
        {"layer", in.Layer},
        {"total", in.Total},
        {"layerReward", in.LayerReward},
        {"layerComputed", in.LayerComputed},
        {"coinbase", in.Coinbase},
        {"smesher", in.Smesher},
    })
    return err
}
