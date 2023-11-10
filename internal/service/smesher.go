package service

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetSmesher returns smesher by id.
func (e *Service) GetSmesher(ctx context.Context, smesherID string) (*model.Smesher, error) {
	smesher, err := e.storage.GetSmesher(ctx, smesherID)
	if err != nil {
		return nil, err
	}
	if smesher == nil {
		return nil, fmt.Errorf("smesher not found `%s`: %w", smesherID, ErrNotFound)
	}
	smesher.Rewards, _, err = e.CountSmesherRewards(ctx, smesherID)
	return smesher, err
}

// GetSmeshers returns smeshers by filter.
func (e *Service) GetSmeshers(ctx context.Context, page, perPage int64) (smeshers []*model.Smesher, total int64, err error) {
	total, err = e.storage.CountSmeshers(ctx, &bson.D{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count total smeshers: %w", err)
	}
	if total == 0 {
		return []*model.Smesher{}, 0, nil
	}
	smeshers, err = e.storage.GetSmeshers(ctx, &bson.D{}, e.getFindOptions("id", page, perPage))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get smeshers: %w", err)
	}
	return smeshers, total, nil
}

// GetSmesherActivations returns smesher activations by filter.
func (e *Service) GetSmesherActivations(ctx context.Context, smesherID string, page, perPage int64) (atxs []*model.Activation, total int64, err error) {
	return e.getActivations(ctx, &bson.D{{Key: "smesher", Value: smesherID}}, e.getFindOptions("layer", page, perPage))
}

// GetSmesherRewards returns smesher rewards by filter.
func (e *Service) GetSmesherRewards(ctx context.Context, smesherID string, page, perPage int64) (rewards []*model.Reward, total int64, err error) {
	opts := e.getFindOptions("layer", page, perPage)
	opts.SetProjection(bson.D{})
	return e.getRewards(ctx, &bson.D{{Key: "smesher", Value: smesherID}}, opts)
}

// CountSmesherRewards returns smesher rewards count by filter.
func (e *Service) CountSmesherRewards(ctx context.Context, smesherID string) (total, count int64, err error) {
	return e.storage.CountSmesherRewards(ctx, smesherID)
}

func (e *Service) getSmeshers(ctx context.Context, filter *bson.D, options *options.FindOptions) (smeshers []*model.Smesher, total int64, err error) {
	atxs, err := e.storage.GetActivations(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error count smeshers: %w", err)
	}

	smeshersList := make([]string, 0, len(atxs))
	var lastID string
	for _, atx := range atxs {
		if lastID != atx.SmesherId {
			smeshersList = append(smeshersList, atx.SmesherId)
			lastID = atx.SmesherId
		}
	}
	total = int64(len(smeshersList))
	if total == 0 {
		return []*model.Smesher{}, 0, nil
	}
	smeshers, err = e.storage.GetSmeshers(ctx, &bson.D{{Key: "id", Value: bson.M{"$in": smeshersList}}})
	if err != nil {
		return nil, 0, fmt.Errorf("error load smeshers: %w", err)
	}
	return smeshers, total, nil
}
