package service

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetReward returns reward by id.
func (e *Service) GetReward(ctx context.Context, rewardID string) (*model.Reward, error) {
	reward, err := e.storage.GetReward(ctx, rewardID)
	if err != nil {
		return nil, fmt.Errorf("error get reward: %w", err)
	}
	if reward == nil {
		return nil, ErrNotFound
	}
	return reward, nil
}

// GetRewards returns rewards by filter.
func (e *Service) GetRewards(ctx context.Context, page, perPage int64) ([]*model.Reward, int64, error) {
	return e.getRewards(ctx, &bson.D{}, options.Find().SetSort(bson.D{{Key: "layer", Value: -1}}).SetLimit(perPage).SetSkip((page-1)*perPage))
}

func (e *Service) getRewards(ctx context.Context, filter *bson.D, options *options.FindOptions) (rewards []*model.Reward, total int64, err error) {
	total, err = e.storage.CountRewards(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error count rewards: %w", err)
	}
	if total == 0 {
		return []*model.Reward{}, 0, nil
	}
	rewards, err = e.storage.GetRewards(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error get rewards: %w", err)
	}
	return rewards, total, nil
}

func (e *Service) GetTotalRewards(ctx context.Context, filter *bson.D) (int64, int64, error) {
	return e.storage.GetTotalRewards(ctx, filter)
}
