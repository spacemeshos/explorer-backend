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

func (s *Storage) InitSmeshersStorage(ctx context.Context) error {
    _, err := s.db.Collection("smeshers").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"id", 1}}, Options: options.Index().SetName("idIndex").SetUnique(true)});
    _, err = s.db.Collection("coinbases").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"address", 1}}, Options: options.Index().SetName("addressIndex").SetUnique(true)});
    return err
}

func (s *Storage) GetSmesher(parent context.Context, query *bson.D) (*model.Smesher, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("smeshers").Find(ctx, query)
    if err != nil {
        log.Info("GetSmesher: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetSmesher: Empty result")
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    smesher := &model.Smesher{
        Id: utils.GetAsString(doc.Lookup("id")),
        Geo: model.Geo{
            utils.GetAsString(doc.Lookup("name")),
            [2]float64 { doc.Lookup("lon").Double(), doc.Lookup("lat").Double() },
        },
        CommitmentSize: utils.GetAsUInt64(doc.Lookup("cSize")),
        Coinbase: utils.GetAsString(doc.Lookup("coinbase")),
        AtxCount: utils.GetAsUInt32(doc.Lookup("atxcount")),
        Timestamp: utils.GetAsUInt32(doc.Lookup("timestamp")),
    }
    return smesher, nil
}

func (s *Storage) GetSmesherByCoinbase(parent context.Context, coinbase string) (*model.Smesher, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("coinbases").Find(ctx, &bson.D{{"address", coinbase}})
    if err != nil {
        log.Info("GetSmesherByCoinbase: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetSmesherByCoinbase: Empty result")
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    smesher := utils.GetAsString(doc.Lookup("smesher"))
    if smesher == "" {
        log.Info("GetSmesherByCoinbase: Empty result")
        return nil, errors.New("Empty result")
    }
    return s.GetSmesher(ctx, &bson.D{{"id", smesher}})
}

func (s *Storage) GetSmeshersCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("smeshers").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetSmeshersCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) IsSmesherExists(parent context.Context, smesher string) bool {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("smeshers").CountDocuments(ctx, bson.D{{"id", smesher}})
    if err != nil {
        log.Info("IsSmesherExists: %v", err)
        return false
    }
    return count > 0
}

func (s *Storage) GetSmeshers(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("smeshers").Find(ctx, query, opts...)
    if err != nil {
        log.Info("GetSmeshers: %v", err)
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        log.Info("GetSmeshers: %v", err)
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        return nil, nil
    }
    return docs.([]bson.D), nil
}

func (s *Storage) SaveSmesher(parent context.Context, in *model.Smesher) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("smeshers").InsertOne(ctx, bson.D{
        {"id", in.Id},
        {"name", in.Geo.Name},
        {"lon", in.Geo.Coordinates[0]},
        {"lat", in.Geo.Coordinates[1]},
        {"cSize", in.CommitmentSize},
        {"coinbase", in.Coinbase},
        {"atxcount", in.AtxCount},
        {"timestamp", in.Timestamp},
    })
    if err != nil {
        log.Info("SaveSmesher: %v", err)
    }
    return err
}

func (s *Storage) UpdateSmesher(parent context.Context, smesher string, coinbase string, space uint64, timestamp uint32) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("coinbases").InsertOne(ctx, bson.D{
        {"address", coinbase},
        {"smesher", smesher},
    })
    if err != nil {
        log.Info("UpdateSmesher: coinbase: %v", err)
    }
    atxCount, err := s.db.Collection("activations").CountDocuments(ctx, &bson.D{{"smesher", smesher}})
    if err != nil {
        log.Info("UpdateSmesher: GetActivationsCount: %v", err)
    }
    _, err = s.db.Collection("smeshers").UpdateOne(ctx, bson.D{{"id", smesher}}, bson.D{
        {"$set", bson.D{
            {"cSize", space},
            {"coinbase", coinbase},
            {"atxcount", atxCount},
            {"timestamp", timestamp},
        }},
    })
    if err != nil {
        log.Info("SaveOrUpdateSmesher: %v", err)
    }
    return err
}
