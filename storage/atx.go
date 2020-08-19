package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"

    "github.com/spacemeshos/explorer-backend/model"
)

func (s *Storage) GetActivation(parent context.Context, query *bson.D) (*model.Activation, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("atxs").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Activation{
        Id: doc.Lookup("id").String(),
        Layer: uint32(doc.Lookup("layer").Int32()),
        SmesherId: doc.Lookup("smesher").String(),
        Coinbase: doc.Lookup("coinbase").String(),
        PrevAtx: doc.Lookup("prevAtx").String(),
        CommitmentSize: uint64(doc.Lookup("cSize").Int64()),
    }
    return account, nil
}

func (s *Storage) GetActivations(parent context.Context, query *bson.D) ([]*model.Activation, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("atxs").Find(ctx, query)
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
    accounts := make([]*model.Activation, len(docs.([]bson.D)), len(docs.([]bson.D)))
    for i, doc := range docs.([]bson.D) {
        accounts[i] = &model.Activation{
            Id: doc[0].Value.(string),
            Layer: uint32(doc[1].Value.(int32)),
            SmesherId: doc[2].Value.(string),
            Coinbase: doc[3].Value.(string),
            PrevAtx: doc[4].Value.(string),
            CommitmentSize: uint64(doc[5].Value.(int64)),
        }
    }
    return accounts, nil
}

func (s *Storage) SaveActivation(parent context.Context, in *model.Activation) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("atxs").InsertOne(ctx, bson.D{
        {"id", in.Id},
        {"layer", in.Layer},
        {"smesher", in.SmesherId},
        {"coinbase", in.Coinbase},
        {"prevAtx", in.PrevAtx},
        {"cSize", in.CommitmentSize},
    })
    return err
}

func (s *Storage) SaveActivations(parent context.Context, in []*model.Activation) error {
    ctx, cancel := context.WithTimeout(parent, 10*time.Second)
    defer cancel()
    for _, atx := range in {
        _, err := s.db.Collection("atxs").InsertOne(ctx, bson.D{
            {"id", atx.Id},
            {"layer", atx.Layer},
            {"smesher", atx.SmesherId},
            {"coinbase", atx.Coinbase},
            {"prevAtx", atx.PrevAtx},
            {"cSize", atx.CommitmentSize},
        })
        if err != nil {
            return err
        }
    }
    return nil
}
