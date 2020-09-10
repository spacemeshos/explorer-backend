package storage

import (
    "context"
    "errors"
    "time"
    "math"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/explorer-backend/model"
    "github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) InitEpochsStorage(ctx context.Context) error {
    _, err := s.db.Collection("epochs").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"number", 1}}, Options: options.Index().SetName("numberIndex").SetUnique(true)});
    return err
}

func (s *Storage) GetEpochByNumber(parent context.Context, epochNumber int32) (*model.Epoch, error) {
    return s.GetEpoch(parent, &bson.D{{"number", epochNumber}})
}

func (s *Storage) GetEpoch(parent context.Context, query *bson.D) (*model.Epoch, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("epochs").Find(ctx, query)
    if err != nil {
        log.Info("GetEpoch: %v", err)
        return nil, err
    }
    if !cursor.Next(ctx) {
        log.Info("GetEpoch: Empty result", err)
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    epoch := &model.Epoch{
        Number: utils.GetAsInt32(doc.Lookup("number")),
        Start: utils.GetAsUInt32(doc.Lookup("start")),
        End: utils.GetAsUInt32(doc.Lookup("end")),
        LayerStart: utils.GetAsUInt32(doc.Lookup("layerstart")),
        LayerEnd: utils.GetAsUInt32(doc.Lookup("layerend")),
        Layers: utils.GetAsUInt32(doc.Lookup("layers")),
    }
    stats := doc.Lookup("stats").Document()
    current := stats.Lookup("current").Document()
    epoch.Stats.Current.Capacity = utils.GetAsInt64(current.Lookup("capacity"))
    epoch.Stats.Current.Decentral = utils.GetAsInt64(current.Lookup("decentral"))
    epoch.Stats.Current.Smeshers = utils.GetAsInt64(current.Lookup("smeshers"))
    epoch.Stats.Current.Transactions = utils.GetAsInt64(current.Lookup("transactions"))
    epoch.Stats.Current.Accounts = utils.GetAsInt64(current.Lookup("accounts"))
    epoch.Stats.Current.Circulation = utils.GetAsInt64(current.Lookup("circulation"))
    epoch.Stats.Current.Rewards = utils.GetAsInt64(current.Lookup("rewards"))
    epoch.Stats.Current.Security = utils.GetAsInt64(current.Lookup("security"))
    epoch.Stats.Current.TxsAmount = utils.GetAsInt64(current.Lookup("txsamount"))
    cumulative := stats.Lookup("cumulative").Document()
    epoch.Stats.Cumulative.Capacity = utils.GetAsInt64(cumulative.Lookup("capacity"))
    epoch.Stats.Cumulative.Decentral = utils.GetAsInt64(cumulative.Lookup("decentral"))
    epoch.Stats.Cumulative.Smeshers = utils.GetAsInt64(cumulative.Lookup("smeshers"))
    epoch.Stats.Cumulative.Transactions = utils.GetAsInt64(cumulative.Lookup("transactions"))
    epoch.Stats.Cumulative.Accounts = utils.GetAsInt64(cumulative.Lookup("accounts"))
    epoch.Stats.Cumulative.Circulation = utils.GetAsInt64(cumulative.Lookup("circulation"))
    epoch.Stats.Cumulative.Rewards = utils.GetAsInt64(cumulative.Lookup("rewards"))
    epoch.Stats.Cumulative.Security = utils.GetAsInt64(cumulative.Lookup("security"))
    epoch.Stats.Cumulative.TxsAmount = utils.GetAsInt64(cumulative.Lookup("txsamount"))
    return epoch, nil
}

func (s *Storage) GetEpochsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("epochs").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetEpochsCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) GetEpochs(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("epochs").Find(ctx, query, opts...)
    if err != nil {
        log.Info("GetEpochs: %v", err)
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        log.Info("GetEpochs: %v", err)
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        log.Info("GetEpochs: Empty result", err)
        return nil, nil
    }
    return docs.([]bson.D), nil
}

func (s *Storage) SaveEpoch(parent context.Context, epoch *model.Epoch) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("epochs").InsertOne(ctx, bson.D{
        {"number", epoch.Number},
        {"start", epoch.Start},
        {"end", epoch.End},
        {"layerstart", epoch.LayerStart},
        {"layerend", epoch.LayerEnd},
        {"layers", epoch.Layers},
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
                {"txsamount", epoch.Stats.Current.TxsAmount},
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
                {"txsamount", epoch.Stats.Cumulative.TxsAmount},
            }},
        }},
    })
    if err != nil {
        log.Info("SaveEpoch: %v", err)
    }
    return err
}

func (s *Storage) SaveOrUpdateEpoch(parent context.Context, epoch *model.Epoch) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    status, err := s.db.Collection("epochs").UpdateOne(ctx, bson.D{{"number", epoch.Number}}, bson.D{
        {"$set", bson.D{
            {"number", epoch.Number},
            {"start", epoch.Start},
            {"end", epoch.End},
            {"layerstart", epoch.LayerStart},
            {"layerend", epoch.LayerEnd},
            {"layers", epoch.Layers},
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
                    {"txsamount", epoch.Stats.Current.TxsAmount},
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
                    {"txsamount", epoch.Stats.Cumulative.TxsAmount},
                }},
            }},
        }},
    }, options.Update().SetUpsert(true))
    if err != nil {
        log.Info("SaveOrUpdateEpoch: %+v, %v", status, err)
    }
    return err
}

func (s *Storage) computeStatistics(epoch *model.Epoch) {
    layerStart, layerEnd := s.GetEpochLayers(epoch.Number)
    if epoch.Start == 0 {
        epoch.LayerStart = layerStart
        epoch.Start = s.NetworkInfo.GenesisTime + layerStart * s.NetworkInfo.LayerDuration
    }
    lastLayer := s.GetLastLayer(context.Background())
    if lastLayer < layerEnd {
        layerEnd = lastLayer
    }
    epoch.LayerEnd = layerEnd
    epoch.End = s.NetworkInfo.GenesisTime + layerEnd * s.NetworkInfo.LayerDuration + s.NetworkInfo.LayerDuration - 1
    epoch.Layers = epoch.LayerEnd - epoch.LayerStart + 1
    duration := float64(s.NetworkInfo.LayerDuration) * float64(s.GetLayersCount(context.Background(), s.GetEpochLayersFilter(epoch.Number, "number")))
    layerFilter := s.GetEpochLayersFilter(epoch.Number, "layer")
    epoch.Stats.Current.Transactions = s.GetTransactionsCount(context.Background(), layerFilter)
    epoch.Stats.Current.TxsAmount = s.GetTransactionsAmount(context.Background(), layerFilter)
    if duration > 0 && s.NetworkInfo.MaxTransactionsPerSecond > 0 {
        epoch.Stats.Current.Capacity = int64(math.Round(((float64(epoch.Stats.Current.Transactions) / duration) / float64(s.NetworkInfo.MaxTransactionsPerSecond)) * 100.0))
    }
    atxs, err := s.GetActivations(context.Background(), layerFilter, options.Find().SetProjection(bson.D{{"_id", 0},{"id", 0},{"layer", 0},{"coinbase", 0},{"prevAtx", 0}}))
    if err != nil {
        return
    }
    smeshers := make(map[string]int64)
    for _, atx  := range atxs {
        var cSize int64
        var smesher string
        for _, e := range atx {
            if e.Key == "smesher" {
                smesher = e.Value.(string)
                continue
            }
            if e.Key == "cSize" {
                if value, ok := atx[1].Value.(int64); ok {
                    cSize = value
                } else if value, ok := atx[1].Value.(int32); ok {
                    cSize = int64(value)
                }
            }
        }
        if smesher != "" {
            smeshers[smesher] += cSize
            epoch.Stats.Current.Security += cSize
        }
    }
    epoch.Stats.Current.Smeshers = int64(len(smeshers))
    // degree_of_decentralization is defined as: 0.5 * (min(n,1e4)^2/1e8) + 0.5 * (1 - gini_coeff(last_100_epochs))
    a := math.Min(float64(epoch.Stats.Current.Smeshers), 1e4)
    epoch.Stats.Current.Decentral = int64(100.0 * (0.5 * (a * a) /1e8  + 0.5 * (1.0 - utils.Gini(smeshers))))
//    for _, account := range epoch.history.accounts {
//        if account.Balance > 0 {
//            stats.accounts++
//            stats.circulation += uint64(account.Balance)
//        }
//    }
    epoch.Stats.Current.Rewards = s.GetLayersRewards(context.Background(), layerStart, layerEnd)
}
