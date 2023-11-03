package storagereader

import (
	"context"
	"fmt"
	"github.com/spacemeshos/explorer-backend/utils"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountTransactions returns the number of transactions matching the query.
func (s *Reader) CountTransactions(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("txs").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count transactions: %w", err)
	}
	return count, nil
}

// GetTransactions returns the transactions matching the query.
func (s *Reader) GetTransactions(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Transaction, error) {
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

func (s *Reader) CountSentTransactions(ctx context.Context, address string) (amount, fees, count int64, err error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "sender", Value: address}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "amount", Value: bson.D{
				{Key: "$sum", Value: "$amount"},
			}},
			{Key: "fees", Value: bson.D{
				{Key: "$sum", Value: "$fee"},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}},
	}
	cursor, err := s.db.Collection("txs").Aggregate(ctx, mongo.Pipeline{
		matchStage,
		groupStage,
	})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error get sent txs: %w", err)
	}
	if !cursor.Next(ctx) {
		return 0, 0, 0, nil
	}
	doc := cursor.Current
	return utils.GetAsInt64(doc.Lookup("amount")),
		utils.GetAsInt64(doc.Lookup("fees")), utils.GetAsInt64(doc.Lookup("count")), nil
}

func (s *Reader) CountReceivedTransactions(ctx context.Context, address string) (amount, count int64, err error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "receiver", Value: address}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "amount", Value: bson.D{
				{Key: "$sum", Value: "$amount"},
			}},
			{Key: "fees", Value: bson.D{
				{Key: "$sum", Value: "$fee"},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}},
	}
	cursor, err := s.db.Collection("txs").Aggregate(ctx, mongo.Pipeline{
		matchStage,
		groupStage,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error get received txs: %w", err)
	}
	if !cursor.Next(ctx) {
		return 0, 0, nil
	}
	doc := cursor.Current
	return utils.GetAsInt64(doc.Lookup("amount")), utils.GetAsInt64(doc.Lookup("count")), nil
}

// GetLatestSentTransaction returns the latest tx for given address
func (s *Reader) GetLatestSentTransaction(ctx context.Context, address string) (*model.Transaction, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "sender", Value: address}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "layer", Value: bson.D{
				{Key: "$max", Value: "$layer"},
			}},
		}},
	}

	cursor, err := s.db.Collection("txs").Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		return nil, fmt.Errorf("error occured while getting latest reward: %w", err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}

	var tx *model.Transaction
	if err = cursor.Decode(&tx); err != nil {
		return nil, fmt.Errorf("error decode reward: %w", err)
	}
	return tx, nil
}
