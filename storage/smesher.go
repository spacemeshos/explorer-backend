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
    account := &model.Smesher{
        Id: doc.Lookup("id").String(),
        Geo: model.Geo{
            doc.Lookup("name").String(),
            [2]float64 { doc.Lookup("lon").Double(), doc.Lookup("lat").Double() },
        },
        CommitmentSize: utils.GetAsUInt64(doc.Lookup("cSize")),
    }
    return account, nil
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
    })
    if err != nil {
        log.Info("SaveSmesher: %v", err)
    }
    return err
}

func (s *Storage) SaveOrUpdateSmesher(parent context.Context, in *model.Smesher) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    status, err := s.db.Collection("smeshers").UpdateOne(ctx, bson.D{{"id", in.Id}}, bson.D{
        {"$set", bson.D{
            {"id", in.Id},
            {"name", in.Geo.Name},
            {"lon", in.Geo.Coordinates[0]},
            {"lat", in.Geo.Coordinates[1]},
            {"cSize", in.CommitmentSize},
        }},
    }, options.Update().SetUpsert(true))
    if err != nil {
        log.Info("SaveOrUpdateSmesher: %+v, %v", status, err)
    }
    return err
}
