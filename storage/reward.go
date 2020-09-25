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

func (s *Storage) InitRewardsStorage(ctx context.Context) error {
    models := []mongo.IndexModel{
        {Keys: bson.D{{"layer", 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)},
        {Keys: bson.D{{"smesher", 1}}, Options: options.Index().SetName("smesherIndex").SetUnique(false)},
        {Keys: bson.D{{"coinbase", 1}}, Options: options.Index().SetName("coinbaseIndex").SetUnique(false)},
        {Keys: bson.D{{"layer", 1}, {"smesher", 1}, {"coinbase", 1}}, Options: options.Index().SetName("rewardIndex").SetUnique(false)},
    }
    _, err := s.db.Collection("rewards").Indexes().CreateMany(ctx, models, options.CreateIndexes().SetMaxTime(20 * time.Second));
    return err
}

func (s *Storage) GetReward(parent context.Context, query *bson.D) (*model.Reward, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("rewards").Find(ctx, query)
    if err != nil {
        log.Info("GetReward: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetReward: Empty result")
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Reward{
        Layer: utils.GetAsUInt32(doc.Lookup("layer")),
        Total: utils.GetAsUInt64(doc.Lookup("total")),
        LayerReward: utils.GetAsUInt64(doc.Lookup("layerReward")),
        LayerComputed: utils.GetAsUInt32(doc.Lookup("layerComputed")),
        Coinbase: utils.GetAsString(doc.Lookup("coinbase")),
        Smesher: utils.GetAsString(doc.Lookup("smesher")),
        Space: utils.GetAsUInt64(doc.Lookup("space")),
        Timestamp: utils.GetAsUInt32(doc.Lookup("timestamp")),
    }
    return account, nil
}

func (s *Storage) GetRewardsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("rewards").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetRewardsCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) GetLayersRewards(parent context.Context, layerStart uint32, layerEnd uint32) (int64, int64) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    matchStage := bson.D{
        {"$match", bson.D{
            {"layer", bson.D{{"$gte", layerStart}, {"$lte", layerEnd}}},
        }},
    }
    groupStage := bson.D{
        {"$group", bson.D{
            {"_id", ""},
            {"total", bson.D{
                {"$sum", "$total"},
            }},
            {"layerReward", bson.D{
                {"$sum", "$layerReward"},
            }},
            {"count", bson.D{
                {"$sum", 1},
            }},
        }},
    }
    cursor, err := s.db.Collection("rewards").Aggregate(ctx, mongo.Pipeline{
        matchStage,
        groupStage,
    })
    if err != nil {
        log.Info("GetLayersRewards: %v", err)
        return 0, 0
    }
    if !cursor.Next(ctx) {
        log.Info("GetLayersRewards: Empty result")
        return 0, 0
    }
    doc := cursor.Current
//    log.Info("LayersRewards(%v, %v): %v + %v", layerStart, layerEnd, utils.GetAsInt64(doc.Lookup("total")), utils.GetAsInt64(doc.Lookup("layerReward")))
//    reward := utils.GetAsInt64(doc.Lookup("total")) + utils.GetAsInt64(doc.Lookup("layerReward"))
    return utils.GetAsInt64(doc.Lookup("total")), utils.GetAsInt64(doc.Lookup("count"))
}

func (s *Storage) GetSmesherRewards(parent context.Context, smesher string) (int64, int64) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    matchStage := bson.D{{"$match", bson.D{{"smesher", smesher}}}}
    groupStage := bson.D{
        {"$group", bson.D{
            {"_id", ""},
            {"total", bson.D{
                {"$sum", "$total"},
            }},
            {"layerReward", bson.D{
                {"$sum", "$layerReward"},
            }},
            {"count", bson.D{
                {"$sum", 1},
            }},
        }},
    }
    cursor, err := s.db.Collection("rewards").Aggregate(ctx, mongo.Pipeline{
        matchStage,
        groupStage,
    })
    if err != nil {
        log.Info("GetSmesherRewards: %v", err)
        return 0, 0
    }
    if !cursor.Next(ctx) {
        log.Info("GetSmesherRewards: Empty result")
        return 0, 0
    }
    doc := cursor.Current
//    log.Info("LayersRewards(%v, %v): %v + %v", layerStart, layerEnd, utils.GetAsInt64(doc.Lookup("total")), utils.GetAsInt64(doc.Lookup("layerReward")))
//    reward := utils.GetAsInt64(doc.Lookup("total")) + utils.GetAsInt64(doc.Lookup("layerReward"))
    return utils.GetAsInt64(doc.Lookup("total")), utils.GetAsInt64(doc.Lookup("count"))
}

func (s *Storage) GetRewards(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("rewards").Find(ctx, query, opts...)
    if err != nil {
        log.Info("GetRewards: %v", err)
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        log.Info("GetRewards: %v", err)
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        return nil, nil
    }
    return docs.([]bson.D), nil
}

func (s *Storage) SaveReward(parent context.Context, in *model.Reward) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("rewards").InsertOne(ctx, bson.D{
        {"layer", in.Layer},
        {"total", in.Total},
        {"layerReward", in.LayerReward},
        {"layerComputed", in.LayerComputed},
        {"coinbase", in.Coinbase},
        {"smesher", in.Smesher},
        {"space", in.Space},
        {"timestamp", in.Timestamp},
    })
    if err != nil {
        log.Info("SaveReward: %v", err)
    }
    return err
}
