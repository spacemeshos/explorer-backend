package service

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetApps returns apps by filter.
func (e *Service) GetApps(ctx context.Context, page, pageSize int64) (apps []*model.App, total int64, err error) {
	return e.getApps(ctx, &bson.D{}, e.getFindOptions("address", page, pageSize))
}

// GetApp returns app by address.
func (e *Service) GetApp(ctx context.Context, appID string) (*model.App, error) {
	filter := &bson.D{{Key: "address", Value: appID}}
	apps, _, err := e.getApps(ctx, filter, options.Find().SetLimit(1).SetProjection(bson.D{{Key: "_id", Value: 0}}))
	if err != nil {
		return nil, fmt.Errorf("error find app: %w", err)
	}
	if len(apps) == 0 {
		return nil, ErrNotFound
	}
	return apps[0], nil
}

func (e *Service) getApps(ctx context.Context, filter *bson.D, options *options.FindOptions) (apps []*model.App, total int64, err error) {
	total, err = e.storage.CountApps(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error count apps: %w", err)
	}
	if total == 0 {
		return []*model.App{}, 0, nil
	}
	apps, err = e.storage.GetApps(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error get apps: %w", err)
	}
	return apps, total, nil
}
