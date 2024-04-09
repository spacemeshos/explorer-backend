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

func (s *Storage) InitActivationsStorage(ctx context.Context) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetName("idIndex").SetUnique(true)},
		{Keys: bson.D{{Key: "layer", Value: 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "smesher", Value: 1}}, Options: options.Index().SetName("smesherIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "coinbase", Value: 1}}, Options: options.Index().SetName("coinbaseIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "targetEpoch", Value: 1}}, Options: options.Index().SetName("targetEpochIndex").SetUnique(false)},
	}
	_, err := s.db.Collection("activations").Indexes().CreateMany(ctx, models, options.CreateIndexes().SetMaxTime(20*time.Second))
	return err
}

func (s *Storage) GetActivation(parent context.Context, query *bson.D) (*model.Activation, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()
	cursor, err := s.db.Collection("activations").Find(ctx, query)
	if err != nil {
		log.Info("GetActivation: %v", err)
		return nil, err
	}
	if !cursor.Next(ctx) {
		log.Info("GetActivation: Empty result")
		return nil, errors.New("empty result")
	}
	doc := cursor.Current
	var received map[string]int64
	err = doc.Lookup("cReceived").Unmarshal(&received)
	if err != nil {
		return nil, err
	}

	account := &model.Activation{
		Id:                utils.GetAsString(doc.Lookup("id")),
		SmesherId:         utils.GetAsString(doc.Lookup("smesher")),
		Coinbase:          utils.GetAsString(doc.Lookup("coinbase")),
		PrevAtx:           utils.GetAsString(doc.Lookup("prevAtx")),
		NumUnits:          utils.GetAsUInt32(doc.Lookup("numunits")),
		CommitmentSize:    utils.GetAsUInt64(doc.Lookup("commitmentSize")),
		PublishEpoch:      utils.GetAsUInt32(doc.Lookup("publishEpoch")),
		TargetEpoch:       utils.GetAsUInt32(doc.Lookup("targetEpoch")),
		Received:          received,
		TickCount:         utils.GetAsUInt64(doc.Lookup("tickCount")),
		Weight:            utils.GetAsUInt64(doc.Lookup("weight")),
		EffectiveNumUnits: utils.GetAsUInt32(doc.Lookup("effectiveNumUnits")),
	}
	return account, nil
}

func (s *Storage) GetActivationsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()
	count, err := s.db.Collection("activations").CountDocuments(ctx, query, opts...)
	if err != nil {
		log.Info("GetActivationsCount: %v", err)
		return 0
	}
	return count
}

func (s *Storage) GetActivations(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()
	cursor, err := s.db.Collection("activations").Find(ctx, query, opts...)
	if err != nil {
		log.Info("GetActivations: %v", err)
		return nil, err
	}
	var docs interface{} = []bson.D{}
	err = cursor.All(ctx, &docs)
	if err != nil {
		log.Info("GetActivations: %v", err)
		return nil, err
	}
	if len(docs.([]bson.D)) == 0 {
		log.Info("GetActivations: Empty result")
		return nil, nil
	}
	return docs.([]bson.D), nil
}

func (s *Storage) SaveActivation(parent context.Context, in *model.Activation) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Minute)
	defer cancel()
	_, err := s.db.Collection("activations").UpdateOne(ctx, bson.D{{Key: "id", Value: in.Id}}, bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "id", Value: in.Id},
			{Key: "smesher", Value: in.SmesherId},
			{Key: "coinbase", Value: in.Coinbase},
			{Key: "prevAtx", Value: in.PrevAtx},
			{Key: "numunits", Value: in.NumUnits},
			{Key: "commitmentSize", Value: int64(in.NumUnits) * int64(s.postUnitSize)},
			{Key: "cReceived", Value: in.Received},
			{Key: "publishEpoch", Value: in.PublishEpoch},
			{Key: "targetEpoch", Value: in.TargetEpoch},
			{Key: "tickCount", Value: in.TickCount},
			{Key: "weight", Value: in.Weight},
			{Key: "effectiveNumUnits", Value: in.EffectiveNumUnits},
		},
	}}, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveActivation: %v", err)
	}
	return err
}

func (s *Storage) SaveOrUpdateActivation(parent context.Context, atx *model.Activation) error {
	filter := bson.D{{Key: "id", Value: atx.Id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "id", Value: atx.Id},
			{Key: "smesher", Value: atx.SmesherId},
			{Key: "coinbase", Value: atx.Coinbase},
			{Key: "prevAtx", Value: atx.PrevAtx},
			{Key: "numunits", Value: atx.NumUnits},
			{Key: "commitmentSize", Value: int64(atx.NumUnits) * int64(s.postUnitSize)},
			{Key: "cReceived." + s.collectorName, Value: atx.Received[s.collectorName]},
			{Key: "publishEpoch", Value: atx.PublishEpoch},
			{Key: "targetEpoch", Value: atx.TargetEpoch},
			{Key: "tickCount", Value: atx.TickCount},
			{Key: "weight", Value: atx.Weight},
			{Key: "effectiveNumUnits", Value: atx.EffectiveNumUnits},
		}},
	}

	_, err := s.db.Collection("activations").UpdateOne(parent, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveOrUpdateActivation: %v", err)
		return err
	}

	return nil
}

func (s *Storage) GetLastActivationReceived() int64 {
	cursor, err := s.db.Collection("activations").Find(context.Background(),
		bson.M{"cReceived." + s.collectorName: bson.M{"$exists": true}},
		options.Find().SetSort(bson.D{{Key: "cReceived." + s.collectorName, Value: -1}}).SetLimit(1))
	if err != nil {
		log.Info("GetLastActivationReceived: %v", err)
		return 0
	}
	if !cursor.Next(context.Background()) {
		log.Info("GetLastActivationReceived: Empty result", err)
		return 0
	}

	doc := cursor.Current
	var received map[string]int64
	err = doc.Lookup("cReceived").Unmarshal(&received)
	if err != nil {
		log.Info("GetLastActivationReceived: %v", err)
		return 0
	}

	receivedTs, ok := received[s.collectorName]
	if !ok {
		return 0
	}

	return receivedTs
}
