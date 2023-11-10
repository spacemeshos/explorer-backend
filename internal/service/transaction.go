package service

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetTransaction returns tx by id.
func (e *Service) GetTransaction(ctx context.Context, txID string) (*model.Transaction, error) {
	filter := &bson.D{{Key: "id", Value: strings.ToLower(txID)}}
	txs, total, err := e.getTransactions(ctx, filter, options.Find().SetLimit(1).SetProjection(bson.D{{Key: "_id", Value: 0}}))
	if err != nil {
		return nil, fmt.Errorf("error get transaction: %w", err)
	}
	if total == 0 {
		return nil, ErrNotFound
	}
	return txs[0], nil
}

// GetTransactions returns txs by filter.
func (e *Service) GetTransactions(ctx context.Context, page, perPage int64) (txs []*model.Transaction, total int64, err error) {
	return e.getTransactions(ctx, &bson.D{}, e.getFindOptions("layer", page, perPage))
}

func (e *Service) getTransactions(ctx context.Context, filter *bson.D, options *options.FindOptions) (txs []*model.Transaction, total int64, err error) {
	total, err = e.storage.CountTransactions(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error count txs: %w", err)
	}
	if total == 0 {
		return []*model.Transaction{}, 0, nil
	}
	txs, err = e.storage.GetTransactions(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error get txs: %w", err)
	}
	return txs, total, nil
}
