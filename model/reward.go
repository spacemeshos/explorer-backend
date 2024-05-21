package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Reward struct {
	ID            string `json:"_id" bson:"_id"`
	Layer         uint32 `json:"layer" bson:"layer"`
	Total         uint64 `json:"total" bson:"total"`
	LayerReward   uint64 `json:"layerReward" bson:"layerReward"`
	LayerComputed uint32 `json:"layerComputed" bson:"layerComputed"` // layer number of the layer when reward was computed
	// tx_fee = total - layer_reward
	Coinbase  string `json:"coinbase" bson:"coinbase"` // account awarded this reward
	Smesher   string `json:"smesher" bson:"smesher"`
	Timestamp uint32 `json:"timestamp" bson:"timestamp"`
}

type RewardService interface {
	GetReward(ctx context.Context, rewardID string) (*Reward, error)
	GetRewardV2(ctx context.Context, smesherID string, layer uint32) (*Reward, error)
	GetRewards(ctx context.Context, page, perPage int64) ([]*Reward, int64, error)
	GetTotalRewards(ctx context.Context, filter *bson.D) (int64, int64, error)
}

func NewReward(reward *pb.Reward) *Reward {
	return &Reward{
		Layer:         reward.GetLayer().GetNumber(),
		Total:         reward.GetTotal().GetValue(),
		LayerReward:   reward.GetLayerReward().GetValue(),
		LayerComputed: reward.GetLayerComputed().GetNumber(),
		Coinbase:      reward.GetCoinbase().GetAddress(),
		Smesher:       utils.BytesToHex(reward.GetSmesher().GetId()),
	}
}
