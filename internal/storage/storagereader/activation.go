package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountActivations returns the number of activations matching the query.
func (s *StorageReader) CountActivations(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("activations").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count activations: %w", err)
	}
	return count, nil
}

// GetActivations returns the activations matching the query.
func (s *StorageReader) GetActivations(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Activation, error) {
	cursor, err := s.db.Collection("activations").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activations: %w", err)
	}
	var docs []*model.Activation
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("error decode activations: %w", err)
	}
	return docs, nil
}
