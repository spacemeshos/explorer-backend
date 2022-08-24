package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountBlocks returns the number of blocks matching the query.
func (s *StorageReader) CountBlocks(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("blocks").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count blocks: %w", err)
	}
	return count, nil
}

// GetBlocks returns the blocks matching the query.
func (s *StorageReader) GetBlocks(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Block, error) {
	cursor, err := s.db.Collection("blocks").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get blocks: %w", err)
	}

	var blocks []*model.Block
	if err = cursor.All(ctx, &blocks); err != nil {
		return nil, fmt.Errorf("error decode blocks: %w", err)
	}
	return blocks, nil
}
