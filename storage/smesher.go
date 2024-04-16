package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spacemeshos/go-spacemesh/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) InitSmeshersStorage(ctx context.Context) error {
	_, err := s.db.Collection("smeshers").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetName("idIndex").SetUnique(true)})
	if err != nil {
		return fmt.Errorf("error init `smeshers` collection: %w", err)
	}
	_, err = s.db.Collection("coinbases").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "smesherId", Value: 1}}, Options: options.Index().SetName("smesherIdIndex").SetUnique(true)})
	if err != nil {
		return fmt.Errorf("error init `coinbases` collection: %w", err)
	}
	return nil
}

func (s *Storage) GetSmesher(parent context.Context, query *bson.D) (*model.Smesher, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("smeshers").Find(ctx, query)
	if err != nil {
		log.Info("GetSmesher: %v", err)
		return nil, err
	}
	if !cursor.Next(ctx) {
		log.Info("GetSmesher: Empty result")
		return nil, errors.New("Empty result")
	}
	doc := cursor.Current
	smesher := &model.Smesher{
		Id:             utils.GetAsString(doc.Lookup("id")),
		CommitmentSize: utils.GetAsUInt64(doc.Lookup("cSize")),
		Coinbase:       utils.GetAsString(doc.Lookup("coinbase")),
		AtxCount:       utils.GetAsUInt32(doc.Lookup("atxcount")),
		Timestamp:      utils.GetAsUInt64(doc.Lookup("timestamp")),
	}
	return smesher, nil
}

func (s *Storage) GetSmeshersCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("smeshers").CountDocuments(ctx, query, opts...)
	if err != nil {
		log.Info("GetSmeshersCount: %v", err)
		return 0
	}
	return count
}

func (s *Storage) IsSmesherExists(parent context.Context, smesher string) bool {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("smeshers").CountDocuments(ctx, bson.D{{Key: "id", Value: smesher}})
	if err != nil {
		log.Info("IsSmesherExists: %v", err)
		return false
	}
	return count > 0
}

func (s *Storage) GetSmeshers(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("smeshers").Find(ctx, query, opts...)
	if err != nil {
		log.Info("GetSmeshers: %v", err)
		return nil, err
	}
	var docs interface{} = []bson.D{}
	err = cursor.All(ctx, &docs)
	if err != nil {
		log.Info("GetSmeshers: %v", err)
		return nil, err
	}
	if len(docs.([]bson.D)) == 0 {
		return nil, nil
	}
	return docs.([]bson.D), nil
}

func (s *Storage) SaveSmesher(parent context.Context, in *model.Smesher, epoch uint32) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("smeshers").UpdateOne(ctx, bson.D{{Key: "id", Value: in.Id}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "id", Value: in.Id},
			{Key: "cSize", Value: in.CommitmentSize},
			{Key: "coinbase", Value: in.Coinbase},
			{Key: "atxcount", Value: in.AtxCount},
			{Key: "timestamp", Value: in.Timestamp},
		}},
		{Key: "$addToSet", Value: bson.M{"epochs": epoch}},
	}, opts)
	if err != nil {
		return fmt.Errorf("error save smesher: %w", err)
	}
	return nil
}

func (s *Storage) SaveSmesherQuery(in *model.Smesher) *mongo.UpdateOneModel {
	filter := bson.D{{Key: "id", Value: in.Id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "id", Value: in.Id},
			{Key: "cSize", Value: in.CommitmentSize},
			{Key: "coinbase", Value: in.Coinbase},
			{Key: "atxcount", Value: in.AtxCount},
			{Key: "timestamp", Value: in.Timestamp},
		}},
	}

	updateModel := mongo.NewUpdateOneModel()
	updateModel.Filter = filter
	updateModel.Update = update
	updateModel.SetUpsert(true)

	return updateModel
}

func (s *Storage) UpdateSmesher(parent context.Context, in *model.Smesher, epoch uint32) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "smesherId", Value: in.Id}}
	_, err := s.db.Collection("coinbases").UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: bson.D{{Key: "coinbase", Value: in.Coinbase}}}},
		options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("error insert smesher into `coinbases`: %w", err)
	}

	atxCount, err := s.db.Collection("activations").CountDocuments(ctx, &bson.D{{Key: "smesher", Value: in.Id}})
	if err != nil {
		log.Info("UpdateSmesher: GetActivationsCount: %v", err)
	}

	_, err = s.db.Collection("smeshers").UpdateOne(ctx, bson.D{{Key: "id", Value: in.Id}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "id", Value: in.Id},
			{Key: "cSize", Value: in.CommitmentSize},
			{Key: "coinbase", Value: in.Coinbase},
			{Key: "timestamp", Value: in.Timestamp},
			{Key: "atxcount", Value: atxCount},
		}},
		{Key: "$addToSet", Value: bson.M{"epochs": epoch}},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("UpdateSmesher: %v", err)
	}
	return err
}

func (s *Storage) UpdateSmesherQuery(in *model.Smesher, epoch uint32) (*mongo.UpdateOneModel, *mongo.UpdateOneModel) {
	coinbaseFilter := bson.D{{Key: "smesherId", Value: in.Id}}
	coinbaseUpdate := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "coinbase", Value: in.Coinbase},
		}}}
	coinbaseModel := mongo.NewUpdateOneModel()
	coinbaseModel.SetFilter(coinbaseFilter)
	coinbaseModel.SetUpdate(coinbaseUpdate)
	coinbaseModel.SetUpsert(true)

	atxCount, err := s.db.Collection("activations").CountDocuments(context.TODO(), &bson.D{{Key: "smesher", Value: in.Id}})
	if err != nil {
		log.Info("UpdateSmesher: GetActivationsCount: %v", err)
	}

	smesherFilter := bson.D{{Key: "id", Value: in.Id}}
	smesherUpdate := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "id", Value: in.Id},
			{Key: "cSize", Value: in.CommitmentSize},
			{Key: "coinbase", Value: in.Coinbase},
			{Key: "timestamp", Value: in.Timestamp},
			{Key: "atxcount", Value: atxCount},
		}},
		{Key: "$addToSet", Value: bson.M{"epochs": epoch}},
	}

	smesherModel := mongo.NewUpdateOneModel()
	smesherModel.SetFilter(smesherFilter)
	smesherModel.SetUpdate(smesherUpdate)

	return coinbaseModel, smesherModel
}
