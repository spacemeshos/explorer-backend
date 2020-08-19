package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"

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
        Layer: uint32(doc.Lookup("layer").Int32()),
        Total: uint64(doc.Lookup("total").Int64()),
        LayerReward: uint64(doc.Lookup("layerReward").Int64()),
        LayerComputed: uint32(doc.Lookup("layerComputed").Int32()),
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
    if len(docs.([]bson.D)) == 0 {
        return nil, nil
    }
    rewards := make([]*model.Reward, len(docs.([]bson.D)), len(docs.([]bson.D)))
    for i, doc := range docs.([]bson.D) {
        rewards[i] = &model.Reward{
            Layer: uint32(doc[0].Value.(int32)),
            Total: uint64(doc[1].Value.(int64)),
            LayerReward: uint64(doc[2].Value.(int64)),
            LayerComputed: uint32(doc[3].Value.(int32)),
            Coinbase: doc[4].Value.(string),
            Smesher: doc[5].Value.(string),
        }
    }
    return rewards, nil
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
