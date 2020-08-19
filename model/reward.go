package model

import (
    "go.mongodb.org/mongo-driver/bson"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Reward struct {
    Layer		uint64
    Total		uint64
    LayerReward		uint64
    LayerComputed	uint64	// layer number of the layer when reward was computed
    // tx_fee = total - layer_reward
    Coinbase		string	// account awarded this reward
    Smesher		string	// it will be nice to always have this in reward events
}

type RewardService interface {
    GetReward(ctx context.Context, query *bson.D) (*Reward, error)
    GetRewards(ctx context.Context, query *bson.D) ([]*Reward, error)
    SaveReward(ctx context.Context, in *Reward) error
}

func NewReward(reward *pb.Reward) *Reward {
    var smesherId SmesherID
    copy(smesherId[:], reward.GetSmesher().GetId())
    return &Reward{
        Layer: LayerID(reward.GetLayer().GetNumber()),
        Total: Amount(reward.GetTotal().GetValue()),
        Layer_reward: Amount(reward.GetLayerReward().GetValue()),
        Layer_computed: LayerID(reward.GetLayerComputed().GetNumber()),
        Coinbase: BytesToAddress(reward.GetCoinbase().GetAddress()),
        Smesher: smesherId,
    }
}
