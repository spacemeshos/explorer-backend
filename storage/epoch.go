package storage

import (
	"context"
	"errors"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) InitEpochsStorage(ctx context.Context) error {
	_, err := s.db.Collection("epochs").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "number", Value: 1}}, Options: options.Index().SetName("numberIndex").SetUnique(true)})
	return err
}

func (s *Storage) GetEpochByNumber(parent context.Context, epochNumber int32) (*model.Epoch, error) {
	return s.GetEpoch(parent, &bson.D{{Key: "number", Value: epochNumber}})
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
		Number:     utils.GetAsInt32(doc.Lookup("number")),
		Start:      utils.GetAsUInt32(doc.Lookup("start")),
		End:        utils.GetAsUInt32(doc.Lookup("end")),
		LayerStart: utils.GetAsUInt32(doc.Lookup("layerstart")),
		LayerEnd:   utils.GetAsUInt32(doc.Lookup("layerend")),
		Layers:     utils.GetAsUInt32(doc.Lookup("layers")),
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
	epoch.Stats.Current.RewardsNumber = utils.GetAsInt64(current.Lookup("rewardsnumber"))
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
	epoch.Stats.Cumulative.RewardsNumber = utils.GetAsInt64(cumulative.Lookup("rewardsnumber"))
	epoch.Stats.Cumulative.Security = utils.GetAsInt64(cumulative.Lookup("security"))
	epoch.Stats.Cumulative.TxsAmount = utils.GetAsInt64(cumulative.Lookup("txsamount"))
	return epoch, nil
}

func (s *Storage) GetEpochsData(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Epoch, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("epochs").Find(ctx, query, opts...)
	if err != nil {
		log.Info("GetEpoch: %v", err)
		return nil, err
	}
	epochs := make([]*model.Epoch, 0)
	for cursor.Next(ctx) {
		doc := cursor.Current
		epoch := &model.Epoch{
			Number:     utils.GetAsInt32(doc.Lookup("number")),
			Start:      utils.GetAsUInt32(doc.Lookup("start")),
			End:        utils.GetAsUInt32(doc.Lookup("end")),
			LayerStart: utils.GetAsUInt32(doc.Lookup("layerstart")),
			LayerEnd:   utils.GetAsUInt32(doc.Lookup("layerend")),
			Layers:     utils.GetAsUInt32(doc.Lookup("layers")),
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
		epoch.Stats.Current.RewardsNumber = utils.GetAsInt64(current.Lookup("rewardsnumber"))
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
		epoch.Stats.Cumulative.RewardsNumber = utils.GetAsInt64(cumulative.Lookup("rewardsnumber"))
		epoch.Stats.Cumulative.Security = utils.GetAsInt64(cumulative.Lookup("security"))
		epoch.Stats.Cumulative.TxsAmount = utils.GetAsInt64(cumulative.Lookup("txsamount"))
		epochs = append(epochs, epoch)
	}
	return epochs, nil
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
		{Key: "number", Value: epoch.Number},
		{Key: "start", Value: epoch.Start},
		{Key: "end", Value: epoch.End},
		{Key: "layerstart", Value: epoch.LayerStart},
		{Key: "layerend", Value: epoch.LayerEnd},
		{Key: "layers", Value: epoch.Layers},
		{Key: "stats", Value: bson.D{
			{Key: "current", Value: bson.D{
				{Key: "capacity", Value: epoch.Stats.Current.Capacity},
				{Key: "decentral", Value: epoch.Stats.Current.Decentral},
				{Key: "smeshers", Value: epoch.Stats.Current.Smeshers},
				{Key: "transactions", Value: epoch.Stats.Current.Transactions},
				{Key: "accounts", Value: epoch.Stats.Current.Accounts},
				{Key: "circulation", Value: epoch.Stats.Current.Circulation},
				{Key: "rewards", Value: epoch.Stats.Current.Rewards},
				{Key: "rewardsnumber", Value: epoch.Stats.Current.RewardsNumber},
				{Key: "security", Value: epoch.Stats.Current.Security},
				{Key: "txsamount", Value: epoch.Stats.Current.TxsAmount},
			}},
			{Key: "cumulative", Value: bson.D{
				{Key: "capacity", Value: epoch.Stats.Cumulative.Capacity},
				{Key: "decentral", Value: epoch.Stats.Cumulative.Decentral},
				{Key: "smeshers", Value: epoch.Stats.Cumulative.Smeshers},
				{Key: "transactions", Value: epoch.Stats.Cumulative.Transactions},
				{Key: "accounts", Value: epoch.Stats.Cumulative.Accounts},
				{Key: "circulation", Value: epoch.Stats.Cumulative.Circulation},
				{Key: "rewards", Value: epoch.Stats.Cumulative.Rewards},
				{Key: "rewardsnumber", Value: epoch.Stats.Cumulative.RewardsNumber},
				{Key: "security", Value: epoch.Stats.Cumulative.Security},
				{Key: "txsamount", Value: epoch.Stats.Cumulative.TxsAmount},
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
	status, err := s.db.Collection("epochs").UpdateOne(ctx, bson.D{{Key: "number", Value: epoch.Number}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "number", Value: epoch.Number},
			{Key: "start", Value: epoch.Start},
			{Key: "end", Value: epoch.End},
			{Key: "layerstart", Value: epoch.LayerStart},
			{Key: "layerend", Value: epoch.LayerEnd},
			{Key: "layers", Value: epoch.Layers},
			{Key: "stats", Value: bson.D{
				{Key: "current", Value: bson.D{
					{Key: "capacity", Value: epoch.Stats.Current.Capacity},
					{Key: "decentral", Value: epoch.Stats.Current.Decentral},
					{Key: "smeshers", Value: epoch.Stats.Current.Smeshers},
					{Key: "transactions", Value: epoch.Stats.Current.Transactions},
					{Key: "accounts", Value: epoch.Stats.Current.Accounts},
					{Key: "circulation", Value: epoch.Stats.Current.Circulation},
					{Key: "rewards", Value: epoch.Stats.Current.Rewards},
					{Key: "rewardsnumber", Value: epoch.Stats.Current.RewardsNumber},
					{Key: "security", Value: epoch.Stats.Current.Security},
					{Key: "txsamount", Value: epoch.Stats.Current.TxsAmount},
				}},
				{Key: "cumulative", Value: bson.D{
					{Key: "capacity", Value: epoch.Stats.Cumulative.Capacity},
					{Key: "decentral", Value: epoch.Stats.Cumulative.Decentral},
					{Key: "smeshers", Value: epoch.Stats.Cumulative.Smeshers},
					{Key: "transactions", Value: epoch.Stats.Cumulative.Transactions},
					{Key: "accounts", Value: epoch.Stats.Cumulative.Accounts},
					{Key: "circulation", Value: epoch.Stats.Cumulative.Circulation},
					{Key: "rewards", Value: epoch.Stats.Cumulative.Rewards},
					{Key: "rewardsnumber", Value: epoch.Stats.Cumulative.RewardsNumber},
					{Key: "security", Value: epoch.Stats.Cumulative.Security},
					{Key: "txsamount", Value: epoch.Stats.Cumulative.TxsAmount},
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
		epoch.Start = s.getLayerTimestamp(layerStart)
	}
	epoch.LayerEnd = layerEnd
	epoch.End = s.getLayerTimestamp(layerEnd) + s.NetworkInfo.LayerDuration - 1
	epoch.Layers = epoch.LayerEnd - epoch.LayerStart + 1
	duration := float64(s.NetworkInfo.LayerDuration) * float64(s.GetLayersCount(context.Background(), s.GetEpochLayersFilter(epoch.Number, "number")))
	layerFilter := s.GetEpochLayersFilter(epoch.Number, "layer")
	epoch.Stats.Current.Transactions = s.GetTransactionsCount(context.Background(), layerFilter)
	epoch.Stats.Current.TxsAmount = s.GetTransactionsAmount(context.Background(), layerFilter)
	if duration > 0 && s.NetworkInfo.MaxTransactionsPerSecond > 0 {
		// todo replace to utils.CalcEpochCapacity
		epoch.Stats.Current.Capacity = int64(math.Round(((float64(epoch.Stats.Current.Transactions) / duration) / float64(s.NetworkInfo.MaxTransactionsPerSecond)) * 100.0))
	}
	atxs, _ := s.GetActivations(context.Background(), layerFilter)
	if atxs != nil {
		smeshers := make(map[string]int64)
		for _, atx := range atxs {
			var commitmentSize int64
			var smesher string
			for _, e := range atx {
				if e.Key == "smesher" {
					smesher, _ = e.Value.(string)
					continue
				}
				if e.Key == "commitmentSize" {
					if value, ok := e.Value.(int64); ok {
						commitmentSize = value
					} else if value, ok := e.Value.(int32); ok {
						commitmentSize = int64(value)
					}
				}
			}
			if smesher != "" {
				smeshers[smesher] += commitmentSize
				epoch.Stats.Current.Security += commitmentSize
			}
		}
		epoch.Stats.Current.Smeshers = int64(len(smeshers))
		// degree_of_decentralization is defined as: 0.5 * (min(n,1e4)^2/1e8) + 0.5 * (1 - gini_coeff(last_100_epochs))
		a := math.Min(float64(epoch.Stats.Current.Smeshers), 1e4)
		// todo replace to utils.CalcDecentralCoefficient
		epoch.Stats.Current.Decentral = int64(100.0 * (0.5*(a*a)/1e8 + 0.5*(1.0-utils.Gini(smeshers))))
	}
	epoch.Stats.Current.Accounts = s.GetAccountsCount(context.Background(), &bson.D{{Key: "created", Value: bson.D{{Key: "$lte", Value: layerEnd}}}})
	epoch.Stats.Cumulative.Circulation, _ = s.GetLayersRewards(context.Background(), 0, layerEnd)
	epoch.Stats.Current.Rewards, epoch.Stats.Current.RewardsNumber = s.GetLayersRewards(context.Background(), layerStart, layerEnd)
}
