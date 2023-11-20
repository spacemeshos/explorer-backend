package storagereader

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// StorageReader is the interface for the storage reader. Providing ReadOnly methods.
type StorageReader interface {
	Ping(ctx context.Context) error
	GetNetworkInfo(ctx context.Context) (*model.NetworkInfo, error)

	CountTransactions(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetTransactions(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Transaction, error)
	CountSentTransactions(ctx context.Context, address string) (amount, fees, count int64, err error)
	CountReceivedTransactions(ctx context.Context, address string) (amount, count int64, err error)
	GetLatestTransaction(ctx context.Context, address string) (*model.Transaction, error)
	GetFirstSentTransaction(ctx context.Context, address string) (*model.Transaction, error)

	CountApps(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetApps(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.App, error)

	CountAccounts(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetAccounts(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Account, error)
	GetAccountSummary(ctx context.Context, address string) (*model.AccountSummary, error)

	CountActivations(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetActivations(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Activation, error)

	CountBlocks(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetBlocks(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Block, error)

	CountEpochs(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetEpochs(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Epoch, error)
	GetEpoch(ctx context.Context, epochNumber int) (*model.Epoch, error)

	CountLayers(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetLayers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Layer, error)
	GetLayer(ctx context.Context, layerNumber int) (*model.Layer, error)

	CountRewards(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	CountCoinbaseRewards(ctx context.Context, coinbase string) (total, count int64, err error)
	GetRewards(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Reward, error)
	GetReward(ctx context.Context, rewardID string) (*model.Reward, error)
	GetLatestReward(ctx context.Context, coinbase string) (*model.Reward, error)
	GetTotalRewards(ctx context.Context) (total, count int64, err error)

	CountSmeshers(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error)
	GetSmeshers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Smesher, error)
	GetSmesher(ctx context.Context, smesherID string) (*model.Smesher, error)
	CountSmesherRewards(ctx context.Context, smesherID string) (total, count int64, err error)
}
