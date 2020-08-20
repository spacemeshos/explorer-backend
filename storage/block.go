package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/explorer-backend/model"
)

type BlockService interface {
    GetBlock(ctx context.Context, query *bson.D) (*model.Block, error)
    GetBlocks(ctx context.Context, query *bson.D) ([]*model.Block, error)
    SaveBlock(ctx context.Context, in *model.Block) error
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
        Layer: uint32(doc.Lookup("layer").Int32()),
    }
    return account, nil
}

func (s *Storage) GetBlocks(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Block, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("blocks").Find(ctx, query, opts...)
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
    blocks := make([]*model.Block, len(docs.([]bson.D)), len(docs.([]bson.D)))
    for i, doc := range docs.([]bson.D) {
        blocks[i] = &model.Block{
            Id: doc[0].Value.(string),
            Layer: uint32(doc[1].Value.(int32)),
        }
    }
    return blocks, nil
}

func (s *Storage) SaveBlock(parent context.Context, in *model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("blocks").InsertOne(ctx, bson.D{
        {"id", in.Id},
        {"layer", in.Layer},
    })
    return err
}

func (s *Storage) SaveBlocks(parent context.Context, in []*model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 10*time.Second)
    defer cancel()
    for _, block := range in {
        _, err := s.db.Collection("blocks").InsertOne(ctx, bson.D{
            {"id", block.Id},
            {"layer", block.Layer},
        })
        if err != nil {
            return err
        }
    }
    return nil
}
