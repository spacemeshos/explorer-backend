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

func (s *Storage) InitLayersStorage(ctx context.Context) error {
    _, err := s.db.Collection("layers").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"number", 1}}, Options: options.Index().SetName("numberIndex").SetUnique(true)});
    return err
}

func (s *Storage) GetLayer(parent context.Context, query *bson.D) (*model.Layer, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("layers").Find(ctx, query)
    if err != nil {
        log.Info("GetLayer: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetLayer: Empty result", err)
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Layer{
        Number: utils.GetAsUInt32(doc.Lookup("number")),
        Status: utils.GetAsInt(doc.Lookup("status")),
        Txs: utils.GetAsUInt32(doc.Lookup("txs")),
        Start: utils.GetAsUInt32(doc.Lookup("start")),
        End: utils.GetAsUInt32(doc.Lookup("end")),
        TxsAmount: utils.GetAsUInt64(doc.Lookup("txsamount")),
        AtxCSize: utils.GetAsUInt64(doc.Lookup("atxssize")),
        Rewards: utils.GetAsUInt64(doc.Lookup("rewards")),
    }
    return account, nil
}

func (s *Storage) GetLayersCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("layers").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetLayersCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) GetLastLayer(parent context.Context) uint32 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("layers").Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(1))
    if err != nil {
        log.Info("GetLastLayer: %v", err)
        return 0
    }
    if !cursor.Next(ctx) {
        log.Info("GetLastLayer: Empty result", err)
        return 0
    }
    doc := cursor.Current
    return utils.GetAsUInt32(doc.Lookup("number"))
}

func (s *Storage) GetLayers(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("layers").Find(ctx, query, opts...)
    if err != nil {
        log.Info("GetLayers: %v", err)
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        log.Info("GetLayers: %v", err)
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        log.Info("GetLayers: Empty result", err)
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
        {"rewards", in.Rewards},
    })
    if err != nil {
        log.Info("SaveLayer: %v", err)
    }
    return err
}

func (s *Storage) SaveOrUpdateLayer(parent context.Context, in *model.Layer) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("layers").UpdateOne(ctx, bson.D{{"number", in.Number}}, bson.D{
        {"$set", bson.D{
            {"number", in.Number},
            {"status", in.Status},
            {"txs", in.Txs},
            {"start", in.Start},
            {"end", in.End},
            {"txsamount", in.TxsAmount},
            {"atxssize", in.AtxCSize},
            {"rewards", in.Rewards},
        }},
    }, options.Update().SetUpsert(true))
    if err != nil {
        log.Info("SaveOrUpdateLayer: %v", err)
    }
    return err
}
