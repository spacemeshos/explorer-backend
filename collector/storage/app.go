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

func (s *Storage) InitAppsStorage(ctx context.Context) error {
	_, err := s.db.Collection("apps").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "address", Value: 1}}, Options: options.Index().SetName("addressIndex").SetUnique(true)})
	return err
}

func (s *Storage) GetApp(parent context.Context, query *bson.D) (*model.App, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("apps").Find(ctx, query)
	if err != nil {
		log.Info("GetApp: %v", err)
		return nil, err
	}
	if !cursor.Next(ctx) {
		log.Info("GetApp: Empty result")
		return nil, errors.New("Empty result")
	}
	doc := cursor.Current
	app := &model.App{
		Address: utils.GetAsString(doc.Lookup("address")),
	}
	return app, nil
}

func (s *Storage) GetAppsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("apps").CountDocuments(ctx, query, opts...)
	if err != nil {
		log.Info("GetAppsCount: %v", err)
		return 0
	}
	return count
}

func (s *Storage) GetApps(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("apps").Find(ctx, query, opts...)
	if err != nil {
		log.Info("GetApps: %v", err)
		return nil, err
	}
	var docs interface{} = []bson.D{}
	err = cursor.All(ctx, &docs)
	if err != nil {
		log.Info("GetApps: %v", err)
		return nil, err
	}
	if len(docs.([]bson.D)) == 0 {
		log.Info("GetApps: Empty result")
		return nil, nil
	}
	return docs.([]bson.D), nil
}

func (s *Storage) SaveApp(parent context.Context, in *model.App) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("apps").InsertOne(ctx, bson.D{
		{Key: "address", Value: in.Address},
	})
	if err != nil {
		log.Info("SaveApp: %v", err)
	}
	return err
}
