package storagereader

import (
	"context"
	"fmt"
	bson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountEpochs returns the number of epochs matching the query.
func (s *Reader) CountEpochs(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("epochs").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count epochs: %w", err)
	}
	return count, nil
}

// GetEpochs returns the epochs matching the query.
func (s *Reader) GetEpochs(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Epoch, error) {
	cursor, err := s.db.Collection("epochs").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get epochs: %w", err)
	}
	var epochs []*model.Epoch
	if err = cursor.All(ctx, &epochs); err != nil {
		return nil, err
	}

	for _, epoch := range epochs {
		total, count, err := s.GetTotalRewards(context.TODO(), &bson.D{{Key: "layer", Value: bson.D{
			{Key: "$gte", Value: epoch.LayerStart}, {Key: "$lte", Value: epoch.LayerEnd}}},
		})
		if err != nil {
			return nil, fmt.Errorf("error get total rewards for epoch %d: %w", epoch.Number, err)
		}

		epoch.Stats.Current.Rewards = total
		epoch.Stats.Current.RewardsNumber = count
		epoch.Stats.Cumulative.Rewards = total
		epoch.Stats.Cumulative.RewardsNumber = count
	}

	return epochs, nil
}

// GetEpoch returns the epoch matching the query.
func (s *Reader) GetEpoch(ctx context.Context, epochNumber int) (*model.Epoch, error) {
	cursor, err := s.db.Collection("epochs").Find(ctx, bson.D{{Key: "number", Value: epochNumber}})
	if err != nil {
		return nil, fmt.Errorf("error get epoch `%d`: %w", epochNumber, err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var epoch *model.Epoch
	if err = cursor.Decode(&epoch); err != nil {
		return nil, fmt.Errorf("error decode epoch `%d`: %w", epochNumber, err)
	}

	total, count, err := s.GetTotalRewards(context.TODO(), &bson.D{{Key: "layer", Value: bson.D{
		{Key: "$gte", Value: epoch.LayerStart}, {Key: "$lte", Value: epoch.LayerEnd}}},
	})

	epoch.Stats.Current.Rewards = total
	epoch.Stats.Current.RewardsNumber = count
	epoch.Stats.Cumulative.Rewards = total
	epoch.Stats.Cumulative.RewardsNumber = count

	return epoch, nil
}
