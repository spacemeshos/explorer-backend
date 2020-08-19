package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"

    "github.com/spacemeshos/explorer-backend/model"
)

type Geo struct {
    Name	string `json:"name"`
    Coordinates	[2]float64 `json:"coordinates"`
}

type Smesher struct {
    Id			string
    Geo			Geo
    CommitmentSize	uint64	// commitment size in bytes
}

func (s *Storage) GetSmesher(parent context.Context, query *bson.D) (*model.Smesher, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("smeshers").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Smesher{
        Id: doc.Lookup("id").String(),
        Geo: model.Geo{
            doc.Lookup("name").String(),
            [2]float64 { doc.Lookup("lon").Double(), doc.Lookup("lat").Double() },
        },
        CommitmentSize: uint64(doc.Lookup("cSize").Int64()),
    }
    return account, nil
}

func (s *Storage) GetSmeshers(parent context.Context, query *bson.D) ([]*model.Smesher, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("smeshers").Find(ctx, query)
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
    smeshers := make([]*model.Smesher, len(docs.([]bson.D)), len(docs.([]bson.D)))
    for i, doc := range docs.([]bson.D) {
        smeshers[i] = &model.Smesher{
            Id: doc[0].Value.(string),
            Geo: model.Geo{
                doc[1].Value.(string),
                [2]float64 { doc[2].Value.(float64), doc[3].Value.(float64) },
            },
            CommitmentSize: uint64(doc[4].Value.(int64)),
        }
    }
    return smeshers, nil
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
    return err
}
