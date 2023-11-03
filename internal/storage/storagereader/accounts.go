package storagereader

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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
	var accSummary model.AccountSummary

	totalRewards, _, err := s.CountCoinbaseRewards(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("error occured while getting sum of rewards: %w", err)
	}
	accSummary.Awards = uint64(totalRewards)

	received, _, err := s.CountReceivedTransactions(ctx, address)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("error occured while getting sum of received txs: %w", err)
		}
	}
	accSummary.Received = uint64(received)

	sent, fees, _, err := s.CountSentTransactions(ctx, address)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("error occured while getting sum of sent txs: %w", err)
		}
	}
	accSummary.Sent = uint64(sent)
	accSummary.Fees = uint64(fees)

	return &accSummary, nil
}
