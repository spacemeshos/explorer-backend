package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountApps returns the number of apps matching the query.
func (s *StorageReader) CountApps(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("apps").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count apps: %w", err)
	}
	return count, nil
}

// GetApps returns the apps matching the query.
func (s *StorageReader) GetApps(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.App, error) {
	cursor, err := s.db.Collection("apps").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}
	var docs []*model.App
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("error decode apps: %w", err)
	}
	return docs, nil
}
