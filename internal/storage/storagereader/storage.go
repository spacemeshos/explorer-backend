package storagereader

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// StorageReader is a wrapper around a mongo client. This client is read-only.
type StorageReader struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewStorageReader creates a new storage reader.
func NewStorageReader(ctx context.Context, dbURL string, dbName string) (*StorageReader, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURL))
	if err != nil {
		return nil, fmt.Errorf("error connect to db: %s", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("error ping to db: %s", err)
	}
	reader := &StorageReader{
		client: client,
		db:     client.Database(dbName),
	}
	return reader, nil
}

// GetNetworkInfo returns the network info matching the query.
func (s *StorageReader) GetNetworkInfo(ctx context.Context) (*model.NetworkInfo, error) {
	cursor, err := s.db.Collection("networkinfo").Find(ctx, bson.D{{"id", 1}})
	if err != nil {
		return nil, fmt.Errorf("error get network info: %s", err)
	}
	if !cursor.Next(ctx) {
		return nil, fmt.Errorf("error get network info: %s", errors.New("empty result"))
	}
	var result model.NetworkInfo
	if err = cursor.Decode(&result); err != nil {
		return nil, fmt.Errorf("error decode network info: %s", err)
	}
	return &result, nil
}

// Ping checks if the database is reachable.
func (s *StorageReader) Ping(ctx context.Context) error {
	if s.client == nil {
		return errors.New("storage not initialized")
	}
	return s.client.Ping(ctx, nil)
}
