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

type Layer struct {
    Number	uint64
    Status	int
}

func (s *Storage) GetLayer(parent context.Context, query *bson.D) (*model.Layer, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("layers").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Layer{
        Number: uint64(doc.Lookup("number").Int64()),
        Status: doc.Lookup("status").Int(),
    }
    return account, nil
}

func (s *Storage) GetLayers(parent context.Context, query *bson.D) ([]*model.Layer, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("layers").Find(ctx, query)
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
    layers := make([]*model.Layer, len(docs), len(docs))
    for i, doc := range docs {
        layers[i] = &model.Layer{
            Number: uint64(doc.Lookup("number").Int64()),
            Status: doc.Lookup("status").Int(),
        }
    }
    return layers, nil
}

func (s *Storage) SaveLayer(parent context.Context, in *model.Layer) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("layers").InsertOne(ctx, bson.D{
        {"number", in.Number},
        {"status", in.Status},
    })
    return err
}
