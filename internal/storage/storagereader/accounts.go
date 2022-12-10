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
	return docs, nil
}

// GetAccountSummary returns the summary of the accounts matching the query. Not all accounts from api have filled this data.
func (s *Reader) GetAccountSummary(ctx context.Context, address string) (*model.AccountSummary, error) {
	matchStage := bson.D{{"$match", bson.D{{"address", address}}}}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", ""},
			{"sent", bson.D{
				{"$sum", "$sent"},
			}},
			{"received", bson.D{
				{"$sum", "$received"},
			}},
			{"awards", bson.D{
				{"$sum", "$reward"},
			}},
			{"fees", bson.D{
				{"$sum", "$fee"},
			}},
			{"layer", bson.D{
				{"$max", "$layer"},
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
