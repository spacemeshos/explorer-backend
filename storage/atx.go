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
	}
	_, err := s.db.Collection("activations").Indexes().CreateMany(ctx, models, options.CreateIndexes().SetMaxTime(20*time.Second))
	return err
}

func (s *Storage) GetActivation(parent context.Context, query *bson.D) (*model.Activation, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
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
	account := &model.Activation{
		Id:             utils.GetAsString(doc.Lookup("id")),
		Layer:          utils.GetAsUInt32(doc.Lookup("layer")),
		SmesherId:      utils.GetAsString(doc.Lookup("smesher")),
		Coinbase:       utils.GetAsString(doc.Lookup("coinbase")),
		PrevAtx:        utils.GetAsString(doc.Lookup("prevAtx")),
		NumUnits:       utils.GetAsUInt32(doc.Lookup("numunits")),
		CommitmentSize: utils.GetAsUInt64(doc.Lookup("commitmentSize")),
		Timestamp:      utils.GetAsUInt32(doc.Lookup("timestamp")),
	}
	return account, nil
}

func (s *Storage) GetActivationsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("activations").CountDocuments(ctx, query, opts...)
	if err != nil {
		log.Info("GetActivationsCount: %v", err)
		return 0
	}
	return count
}

func (s *Storage) GetActivations(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
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
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("activations").InsertOne(ctx, bson.D{
		{Key: "id", Value: in.Id},
		{Key: "layer", Value: in.Layer},
		{Key: "smesher", Value: in.SmesherId},
		{Key: "coinbase", Value: in.Coinbase},
		{Key: "prevAtx", Value: in.PrevAtx},
		{Key: "numunits", Value: in.NumUnits},
		{Key: "commitmentSize", Value: int64(in.NumUnits) * int64(s.postUnitSize)},
		{Key: "timestamp", Value: in.Timestamp},
	})
	if err != nil {
		log.Info("SaveActivation: %v", err)
	}
	return err
}

func (s *Storage) SaveActivations(parent context.Context, in []*model.Activation) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	for _, atx := range in {
		_, err := s.db.Collection("activations").InsertOne(ctx, bson.D{
			{Key: "id", Value: atx.Id},
			{Key: "layer", Value: atx.Layer},
			{Key: "smesher", Value: atx.SmesherId},
			{Key: "coinbase", Value: atx.Coinbase},
			{Key: "prevAtx", Value: atx.PrevAtx},
			{Key: "numunits", Value: atx.NumUnits},
			{Key: "commitmentSize", Value: int64(atx.NumUnits) * int64(s.postUnitSize)},
			{Key: "timestamp", Value: atx.Timestamp},
		})
		if err != nil {
			log.Info("SaveActivations: %v", err)
			return err
		}
	}
	return nil
}

func (s *Storage) SaveOrUpdateActivations(parent context.Context, in []*model.Activation) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	for _, atx := range in {
		_, err := s.db.Collection("activations").UpdateOne(ctx, bson.D{{Key: "id", Value: atx.Id}}, bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "id", Value: atx.Id},
				{Key: "layer", Value: atx.Layer},
				{Key: "smesher", Value: atx.SmesherId},
				{Key: "coinbase", Value: atx.Coinbase},
				{Key: "prevAtx", Value: atx.PrevAtx},
				{Key: "numunits", Value: atx.NumUnits},
				{Key: "commitmentSize", Value: int64(atx.NumUnits) * int64(s.postUnitSize)},
				{Key: "timestamp", Value: atx.Timestamp},
			}},
		}, options.Update().SetUpsert(true))
		if err != nil {
			log.Info("SaveOrUpdateActivations: %v", err)
			return err
		}
	}
	return nil
}
