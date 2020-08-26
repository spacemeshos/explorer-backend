package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/explorer-backend/model"
)

func (s *Storage) InitEpochsStorage(ctx context.Context) error {
    _, err := s.db.Collection("epochs").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"number", 1}}, Options: options.Index().SetName("numberIndex").SetUnique(true)});
    return err
}

func (s *Storage) GetEpoch(parent context.Context, query *bson.D) (*model.Epoch, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("epochs").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    epoch := &model.Epoch{
        Number: doc.Lookup("number").Int32(),
    }
    stats := doc.Lookup("stats").Document()
    current := stats.Lookup("current").Document()
    epoch.Stats.Current.Capacity = uint64(current.Lookup("capacity").Int64())
    epoch.Stats.Current.Decentral = uint64(current.Lookup("decentral").Int64())
    epoch.Stats.Current.Smeshers = uint64(current.Lookup("smeshers").Int64())
    epoch.Stats.Current.Transactions = uint64(current.Lookup("transactions").Int64())
    epoch.Stats.Current.Accounts = uint64(current.Lookup("accounts").Int64())
    epoch.Stats.Current.Circulation = uint64(current.Lookup("circulation").Int64())
    epoch.Stats.Current.Rewards = uint64(current.Lookup("rewards").Int64())
    epoch.Stats.Current.Security = uint64(current.Lookup("security").Int64())
    cumulative := stats.Lookup("cumulative").Document()
    epoch.Stats.Cumulative.Capacity = uint64(cumulative.Lookup("capacity").Int64())
    epoch.Stats.Cumulative.Decentral = uint64(cumulative.Lookup("decentral").Int64())
    epoch.Stats.Cumulative.Smeshers = uint64(cumulative.Lookup("smeshers").Int64())
    epoch.Stats.Cumulative.Transactions = uint64(cumulative.Lookup("transactions").Int64())
    epoch.Stats.Cumulative.Accounts = uint64(cumulative.Lookup("accounts").Int64())
    epoch.Stats.Cumulative.Circulation = uint64(cumulative.Lookup("circulation").Int64())
    epoch.Stats.Cumulative.Rewards = uint64(cumulative.Lookup("rewards").Int64())
    epoch.Stats.Cumulative.Security = uint64(cumulative.Lookup("security").Int64())
    return epoch, nil
}

func (s *Storage) GetEpochsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    return s.db.Collection("epochs").CountDocuments(ctx, query, opts...)
}

func (s *Storage) GetEpochs(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("epochs").Find(ctx, query, opts...)
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
    return docs.([]bson.D), nil
}

func (s *Storage) SaveEpoch(parent context.Context, epoch *model.Epoch) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("epochs").InsertOne(ctx, bson.D{
        {"number", epoch.Number},
        {"stats", bson.D{
            {"current",  bson.D{
                {"capacity", epoch.Stats.Current.Capacity},
                {"decentral", epoch.Stats.Current.Decentral},
                {"smeshers", epoch.Stats.Current.Smeshers},
                {"transactions", epoch.Stats.Current.Transactions},
                {"accounts", epoch.Stats.Current.Accounts},
                {"circulation", epoch.Stats.Current.Circulation},
                {"rewards", epoch.Stats.Current.Rewards},
                {"security", epoch.Stats.Current.Security},
            }},
            {"cumulative",  bson.D{
                {"capacity", epoch.Stats.Cumulative.Capacity},
                {"decentral", epoch.Stats.Cumulative.Decentral},
                {"smeshers", epoch.Stats.Cumulative.Smeshers},
                {"transactions", epoch.Stats.Cumulative.Transactions},
                {"accounts", epoch.Stats.Cumulative.Accounts},
                {"circulation", epoch.Stats.Cumulative.Circulation},
                {"rewards", epoch.Stats.Cumulative.Rewards},
                {"security", epoch.Stats.Cumulative.Security},
            }},
        }},
    })
    return err
}
