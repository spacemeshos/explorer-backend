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

type BlockService interface {
    GetBlock(ctx context.Context, query *bson.D) (*Block, error)
    GetBlocks(ctx context.Context, query *bson.D) ([]*Block, error)
    SaveBlock(ctx context.Context, in *Block) error
}
func (s *Storage) GetBlock(parent context.Context, query *bson.D) (*model.Block, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("blocks").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Block{
        Id: doc.Lookup("id").String(),
        Layer: uint64(doc.Lookup("layer").Int64()),
    }
    return account, nil
}

func (s *Storage) GetBlocks(parent context.Context, query *bson.D) ([]*model.Block, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("blocks").Find(ctx, query)
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
    blocks := make([]*model.Block, len(docs), len(docs))
    for i, doc := range docs {
        blocks[i] = &model.Block{
            Id: doc.Lookup("id").String(),
            Layer: uint64(doc.Lookup("layer").Int64()),
        }
    }
    return blocks, nil
}

func (s *Storage) SaveBlock(parent context.Context, in *model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("blocks").InsertOne(ctx, bson.D{
        {"id", in.Id},
        {"layer", in.Layer},
    })
    return err
}

func (s *Storage) SaveBlocks(parent context.Context, in []*model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 10*time.Second)
    defer cancel()
    for _, block := range in {
        res, err := s.db.Collection("blocks").InsertOne(ctx, bson.D{
            {"id", block.Id},
            {"layer", block.Layer},
        })
        if err != nil {
            return err
        }
    }
    return nil
}
