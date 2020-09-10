package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/explorer-backend/model"
    "github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) GetNetworkInfo(parent context.Context) (*model.NetworkInfo, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("networkinfo").Find(ctx, bson.D{{"id", 1}})
    if err != nil {
        log.Info("GetNetworkInfo: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetNetworkInfo: Empty result")
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    info := &model.NetworkInfo{
        NetId: utils.GetAsUInt32(doc.Lookup("netid")),
        GenesisTime: utils.GetAsUInt32(doc.Lookup("genesis")),
        EpochNumLayers: utils.GetAsUInt32(doc.Lookup("layers")),
        MaxTransactionsPerSecond: utils.GetAsUInt32(doc.Lookup("maxtx")),
        LayerDuration: utils.GetAsUInt32(doc.Lookup("duration")),
    }
    return info, nil
}

func (s *Storage) SaveOrUpdateNetworkInfo(parent context.Context, in *model.NetworkInfo) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("networkinfo").UpdateOne(ctx, bson.D{{"id", 1}}, bson.D{
        {"$set", bson.D{
            {"id", 1},
            {"netid", in.NetId},
            {"genesis", in.GenesisTime},
            {"layers", in.EpochNumLayers},
            {"maxtx", in.MaxTransactionsPerSecond},
            {"duration", in.LayerDuration},
        }},
    }, options.Update().SetUpsert(true))
    if err != nil {
        log.Info("SaveOrUpdateNetworkInfo: %v", err)
    }
    return err
}
