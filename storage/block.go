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

func (s *Storage) InitBlocksStorage(ctx context.Context) error {
    _, err := s.db.Collection("blocks").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"id", 1}}, Options: options.Index().SetName("idIndex").SetUnique(true)});
    return err
}

func (s *Storage) GetBlock(parent context.Context, query *bson.D) (*model.Block, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("blocks").Find(ctx, query)
    if err != nil {
        log.Info("GetBlock: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetBlock: Empty result", err)
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Block{
        Id: utils.GetAsString(doc.Lookup("id")),
        Layer: utils.GetAsUInt32(doc.Lookup("layer")),
        Epoch: utils.GetAsUInt32(doc.Lookup("epoch")),
        Start: utils.GetAsUInt32(doc.Lookup("start")),
        End: utils.GetAsUInt32(doc.Lookup("end")),
        TxsNumber: utils.GetAsUInt32(doc.Lookup("txsnumber")),
        TxsValue: utils.GetAsUInt64(doc.Lookup("txsvalue")),
    }
    return account, nil
}

func (s *Storage) GetBlocksCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("blocks").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetBlocksCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) GetBlocks(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("blocks").Find(ctx, query, opts...)
    if err != nil {
        log.Info("GetBlocks: %v", err)
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        log.Info("GetBlocks: %v", err)
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        log.Info("GetBlocks: Empty result")
        return nil, nil
    }
    return docs.([]bson.D), nil
}

func (s *Storage) SaveBlock(parent context.Context, in *model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("blocks").InsertOne(ctx, bson.D{
        {"id", in.Id},
        {"layer", in.Layer},
        {"epoch", in.Epoch},
        {"start", in.Start},
        {"end", in.End},
        {"txsnumber", in.TxsNumber},
        {"txsvalue", in.TxsValue},
    })
    if err != nil {
        log.Info("SaveBlock: %v", err)
    }
    return err
}

func (s *Storage) SaveBlocks(parent context.Context, in []*model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 10*time.Second)
    defer cancel()
    for _, block := range in {
        _, err := s.db.Collection("blocks").InsertOne(ctx, bson.D{
            {"id", block.Id},
            {"layer", block.Layer},
            {"epoch", block.Epoch},
            {"start", block.Start},
            {"end", block.End},
            {"txsnumber", block.TxsNumber},
            {"txsvalue", block.TxsValue},
        })
        if err != nil {
            log.Info("SaveBlocks: %v", err)
            return err
        }
    }
    return nil
}

func (s *Storage) SaveOrUpdateBlocks(parent context.Context, in []*model.Block) error {
    ctx, cancel := context.WithTimeout(parent, 10*time.Second)
    defer cancel()
    for _, block := range in {
        _, err := s.db.Collection("blocks").UpdateOne(ctx, bson.D{{"id", block.Id}}, bson.D{
            {"$set", bson.D{
                {"id", block.Id},
                {"layer", block.Layer},
                {"epoch", block.Epoch},
                {"start", block.Start},
                {"end", block.End},
                {"txsnumber", block.TxsNumber},
                {"txsvalue", block.TxsValue},
            }},
        }, options.Update().SetUpsert(true))
        if err != nil {
            log.Info("SaveOrUpdateBlocks: %v", err)
            return err
        }
    }
    return nil
}
