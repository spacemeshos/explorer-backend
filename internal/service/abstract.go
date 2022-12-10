package service

import (
	"context"
	"errors"

	"github.com/spacemeshos/explorer-backend/model"
)

// ErrNotFound is returned when a resource is not found. Router will serve 404 error if this is returned.
var ErrNotFound = errors.New("not found")

// AppService is an interface for interacting with the app collection.
type AppService interface {
	GetState(ctx context.Context) (*model.NetworkInfo, *model.Epoch, *model.Layer, error)
	GetNetworkInfo(ctx context.Context) (*model.NetworkInfo, error)
	Search(ctx context.Context, search string) (string, error)
	Ping(ctx context.Context) error

	model.EpochService
	model.LayerService
	model.SmesherService
	model.AccountService
	model.RewardService
	model.TransactionService
	model.ActivationService
	model.AppService
	model.BlockService
}
