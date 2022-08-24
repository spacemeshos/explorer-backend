package service

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetAccount returns account by id.
func (e *Service) GetAccount(ctx context.Context, accountID string) (*model.Account, error) {
	idStr := model.ToCheckedAddress(strings.ToLower(accountID))
	if idStr == "" {
		return nil, ErrNotFound
	}

	filter := &bson.D{{"address", idStr}}
	accs, total, err := e.getAccounts(ctx, filter, options.Find().SetSort(bson.D{{"address", 1}}).SetLimit(1).SetProjection(bson.D{
		{"_id", 0},
		{"layer", 0},
	}))
	if err != nil {
		return nil, fmt.Errorf("error find account: %w", err)
	}
	if total == 0 {
		return nil, ErrNotFound
	}
	acc := accs[0]
	summary, err := e.storage.GetAccountSummary(ctx, acc.Address)
	if err != nil {
		return nil, fmt.Errorf("error get account summary: %w", err)
	}

	if summary != nil {
		acc.Sent = summary.Sent
		acc.Received = summary.Received
		acc.Awards = summary.Awards
		acc.Fees = summary.Fees
		acc.LayerTms = summary.LayerTms
	}

	if acc.LayerTms == 0 {
		net, err := e.GetNetworkInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("error get network info for acc summury: %w", err)
		}
		acc.LayerTms = int32(net.GenesisTime)
	}

	acc.Txs, err = e.storage.CountTransactions(ctx, &bson.D{
		{"$or", bson.A{
			bson.D{{"sender", acc.Address}},
			bson.D{{"receiver", acc.Address}},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("error count transactions: %w", err)
	}
	return acc, nil
}

// GetAccounts returns accounts by filter.
func (e *Service) GetAccounts(ctx context.Context, page, perPage int64) ([]*model.Account, int64, error) {
	return e.getAccounts(ctx, &bson.D{}, e.getFindOptions("layer", page, perPage).SetProjection(bson.D{
		{"_id", 0},
		{"layer", 0},
	}))
}

// GetAccountTransactions returns transactions by account id.
func (e *Service) GetAccountTransactions(ctx context.Context, accountID string, page, perPage int64) ([]*model.Transaction, int64, error) {
	idStr := model.ToCheckedAddress(strings.ToLower(accountID))
	if idStr == "" {
		return nil, 0, ErrNotFound
	}

	filter := &bson.D{
		{"$or", bson.A{
			bson.D{{"sender", idStr}},
			bson.D{{"receiver", idStr}},
		}},
	}
	return e.getTransactions(ctx, filter, e.getFindOptions("counter", page, perPage))
}

// GetAccountRewards returns rewards by account id.
func (e *Service) GetAccountRewards(ctx context.Context, accountID string, page, perPage int64) ([]*model.Reward, int64, error) {
	idStr := model.ToCheckedAddress(strings.ToLower(accountID))
	if idStr == "" {
		return nil, 0, ErrNotFound
	}
	return e.getRewards(ctx, &bson.D{{"coinbase", idStr}}, e.getFindOptions("coinbase", page, perPage))
}

func (e *Service) getAccounts(ctx context.Context, filter *bson.D, options *options.FindOptions) (accs []*model.Account, total int64, err error) {
	total, err = e.storage.CountAccounts(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error count accounts: %w", err)
	}
	if total == 0 {
		return []*model.Account{}, 0, nil
	}
	accs, err = e.storage.GetAccounts(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error get accounts: %w", err)
	}
	return accs, total, nil
}
