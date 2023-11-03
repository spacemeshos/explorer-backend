package storagereader

import (
	"context"
	"fmt"
	"github.com/spacemeshos/explorer-backend/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountRewards returns the number of rewards matching the query.
func (s *Reader) CountRewards(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("rewards").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count transactions: %w", err)
	}
	return count, nil
}

// GetRewards returns the rewards matching the query.
func (s *Reader) GetRewards(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Reward, error) {
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
func (s *Reader) GetReward(ctx context.Context, rewardID string) (*model.Reward, error) {
	id, err := primitive.ObjectIDFromHex(strings.ToLower(rewardID))
	if err != nil {
		return nil, fmt.Errorf("error create objectID from string `%s`: %w", rewardID, err)
	}
	cursor, err := s.db.Collection("rewards").Find(ctx, &bson.D{{Key: "_id", Value: id}})
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

// CountCoinbaseRewards returns the number of rewards for given coinbase address.
func (s *Reader) CountCoinbaseRewards(ctx context.Context, coinbase string) (total, count int64, err error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "coinbase", Value: coinbase}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "total", Value: bson.D{
				{Key: "$sum", Value: "$total"},
			}},
			{Key: "layerReward", Value: bson.D{
				{Key: "$sum", Value: "$layerReward"},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}},
	}
	cursor, err := s.db.Collection("rewards").Aggregate(ctx, mongo.Pipeline{
		matchStage,
		groupStage,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error get coinbase rewards: %w", err)
	}
	if !cursor.Next(ctx) {
		return 0, 0, nil
	}
	doc := cursor.Current
	return utils.GetAsInt64(doc.Lookup("total")), utils.GetAsInt64(doc.Lookup("count")), nil
}

// GetLatestReward returns the latest reward for given coinbase
func (s *Reader) GetLatestReward(ctx context.Context, coinbase string) (*model.Reward, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "coinbase", Value: coinbase}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "layer", Value: bson.D{
				{Key: "$max", Value: "$layer"},
			}},
		}},
	}

	cursor, err := s.db.Collection("rewards").Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		return nil, fmt.Errorf("error occured while getting latest reward: %w", err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}

	var reward *model.Reward
	if err = cursor.Decode(&reward); err != nil {
		return nil, fmt.Errorf("error decode reward: %w", err)
	}
	return reward, nil
}
