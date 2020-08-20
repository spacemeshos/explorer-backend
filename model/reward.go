package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Reward struct {
    Layer		uint32
    Total		uint64
    LayerReward		uint64
    LayerComputed	uint32	// layer number of the layer when reward was computed
    // tx_fee = total - layer_reward
    Coinbase		string	// account awarded this reward
    Smesher		string	// it will be nice to always have this in reward events
}

type RewardService interface {
    GetReward(ctx context.Context, query *bson.D) (*Reward, error)
    GetRewards(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Reward, error)
    SaveReward(ctx context.Context, in *Reward) error
}

func NewReward(reward *pb.Reward) *Reward {
    return &Reward{
        Layer: reward.GetLayer().GetNumber(),
        Total: reward.GetTotal().GetValue(),
        LayerReward: reward.GetLayerReward().GetValue(),
        LayerComputed: reward.GetLayerComputed().GetNumber(),
        Coinbase: utils.BytesToAddressString(reward.GetCoinbase().GetAddress()),
        Smesher: utils.BytesToHex(reward.GetSmesher().GetId()),
    }
}
