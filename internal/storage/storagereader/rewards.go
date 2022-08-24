package storagereader

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountRewards returns the number of rewards matching the query.
func (s *StorageReader) CountRewards(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("rewards").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count transactions: %w", err)
	}
	return count, nil
}

// GetRewards returns the rewards matching the query.
func (s *StorageReader) GetRewards(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Reward, error) {
	cursor, err := s.db.Collection("rewards").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get rewards: %w", err)
	}

	var rewards []*model.Reward
	if err = cursor.All(ctx, &rewards); err != nil {
		return nil, fmt.Errorf("error decode rewards: %w", err)
	}
	return rewards, nil
}

// GetReward returns the reward matching the query.
func (s *StorageReader) GetReward(ctx context.Context, rewardID string) (*model.Reward, error) {
	id, err := primitive.ObjectIDFromHex(strings.ToLower(rewardID))
	if err != nil {
		return nil, fmt.Errorf("error create objectID from string `%s`: %w", rewardID, err)
	}
	cursor, err := s.db.Collection("rewards").Find(ctx, &bson.D{{"_id", id}})
	if err != nil {
		return nil, fmt.Errorf("error get reward `%s`: %w", rewardID, err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var reward *model.Reward
	if err = cursor.Decode(&reward); err != nil {
		return nil, fmt.Errorf("error decode reward `%s`: %w", rewardID, err)
	}
	return reward, nil
}
