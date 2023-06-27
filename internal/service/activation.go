package service

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetActivations returns atxs by filter.
func (e *Service) GetActivations(ctx context.Context, page, perPage int64) (atxs []*model.Activation, total int64, err error) {
	return e.getActivations(ctx, &bson.D{}, e.getFindOptions("layer", page, perPage))
}

// GetActivation returns atx by id.
func (e *Service) GetActivation(ctx context.Context, activationID string) (*model.Activation, error) {
	filter := &bson.D{{Key: "id", Value: strings.ToLower(activationID)}}
	atx, total, err := e.getActivations(ctx, filter, options.Find().SetLimit(1).SetProjection(bson.D{{Key: "_id", Value: 0}}))
	if err != nil {
		return nil, fmt.Errorf("error find atx: %w", err)
	}
	if total == 0 {
		return nil, ErrNotFound
	}
	return atx[0], nil
}

func (e *Service) getActivations(ctx context.Context, filter *bson.D, options *options.FindOptions) (atxs []*model.Activation, total int64, err error) {
	total, err = e.storage.CountActivations(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error count atxs: %w", err)
	}
	if total == 0 {
		return []*model.Activation{}, 0, nil
	}

	atxs, err = e.storage.GetActivations(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error get atxs: %w", err)
	}
	return atxs, total, nil
}
