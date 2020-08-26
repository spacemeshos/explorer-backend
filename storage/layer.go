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

func (s *Storage) InitLayersStorage(ctx context.Context) error {
    _, err := s.db.Collection("layers").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"number", 1}}, Options: options.Index().SetName("numberIndex").SetUnique(true)});
    return err
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
        Number: uint32(doc.Lookup("number").Int32()),
        Status: int(doc.Lookup("status").Int32()),
        Txs: uint32(doc.Lookup("txs").Int32()),
        Start: uint32(doc.Lookup("start").Int32()),
        End: uint32(doc.Lookup("end").Int32()),
        TxsAmount: uint64(doc.Lookup("txsamount").Int64()),
        AtxCSize: uint64(doc.Lookup("atxssize").Int64()),
    }
    return account, nil
}

func (s *Storage) GetLayersCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    return s.db.Collection("layers").CountDocuments(ctx, query, opts...)
}

func (s *Storage) GetLayers(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("layers").Find(ctx, query, opts...)
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

func (s *Storage) SaveLayer(parent context.Context, in *model.Layer) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("layers").InsertOne(ctx, bson.D{
        {"number", in.Number},
        {"status", in.Status},
        {"txs", in.Txs},
        {"start", in.Start},
        {"end", in.End},
        {"txsamount", in.TxsAmount},
        {"atxssize", in.AtxCSize},
    })
    return err
}
