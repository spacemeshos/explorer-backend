package storagereader

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountAccounts returns the number of accounts matching the query.
func (s *Reader) CountAccounts(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	return s.db.Collection("accounts").CountDocuments(ctx, query, opts...)
}

// GetAccounts returns the accounts matching the query.
func (s *Reader) GetAccounts(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Account, error) {
	cursor, err := s.db.Collection("accounts").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	var docs []*model.Account
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("error decode accounts: %w", err)
	}

	for _, doc := range docs {
		summary, err := s.GetAccountSummary(ctx, doc.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to get account summary: %w", err)
		}
		if summary == nil {
			continue
		}

		doc.Sent = summary.Sent
		doc.Received = summary.Received
		doc.Awards = summary.Awards
		doc.Fees = summary.Fees
		doc.LayerTms = int32(s.GetLayerTimestamp(uint32(doc.Created)))
	}
	return docs, nil
}

// GetAccountSummary returns the summary of the accounts matching the query. Not all accounts from api have filled this data.
func (s *Reader) GetAccountSummary(ctx context.Context, address string) (*model.AccountSummary, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "address", Value: address}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "sent", Value: bson.D{
				{Key: "$sum", Value: "$sent"},
			}},
			{Key: "received", Value: bson.D{
				{Key: "$sum", Value: "$received"},
			}},
			{Key: "awards", Value: bson.D{
				{Key: "$sum", Value: "$reward"},
			}},
			{Key: "fees", Value: bson.D{
				{Key: "$sum", Value: "$fee"},
			}},
			{Key: "layer", Value: bson.D{
				{Key: "$max", Value: "$layer"},
			}},
		}},
	}

	cursor, err := s.db.Collection("ledger").Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		return nil, fmt.Errorf("error get account summary: %w", err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var accSummary model.AccountSummary
	if err = cursor.Decode(&accSummary); err != nil {
		return nil, fmt.Errorf("error decode account summary: %w", err)
	}
	return &accSummary, nil
}
