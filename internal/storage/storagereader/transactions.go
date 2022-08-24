package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountTransactions returns the number of transactions matching the query.
func (s *StorageReader) CountTransactions(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("txs").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count transactions: %w", err)
	}
	return count, nil
}

// GetTransactions returns the transactions matching the query.
func (s *StorageReader) GetTransactions(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Transaction, error) {
	cursor, err := s.db.Collection("txs").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get txs: %w", err)
	}

	var txs []*model.Transaction
	if err = cursor.All(ctx, &txs); err != nil {
		return nil, fmt.Errorf("error decode txs: %w", err)
	}
	return txs, nil
}
