package model

import (
	"context"
)

type Statistics struct {
	Capacity      int64 `json:"capacity" bson:"capacity"`         // Average tx/s rate over capacity considering all layers in the current epoch.
	Decentral     int64 `json:"decentral" bson:"decentral"`       // Distribution of storage between all active smeshers.
	Smeshers      int64 `json:"smeshers" bson:"smeshers"`         // Number of active smeshers in the current epoch.
	Transactions  int64 `json:"transactions" bson:"transactions"` // Total number of transactions processed by the state transition function.
	Accounts      int64 `json:"accounts" bson:"accounts"`         // Total number of on-mesh accounts with a non-zero coin balance as of the current epoch.
	Circulation   int64 `json:"circulation" bson:"circulation"`   // Total number of Smesh coins in circulation. This is the total balances of all on-mesh accounts.
	Rewards       int64 `json:"rewards" bson:"rewards"`           // Total amount of Smesh minted as mining rewards as of the last known reward distribution event.
	RewardsNumber int64 `json:"rewardsnumber" bson:"rewardsnumber"`
	Security      int64 `json:"security" bson:"security"`   // Total amount of storage committed to the network based on the ATXs in the previous epoch.
	TxsAmount     int64 `json:"txsamount" bson:"txsamount"` // Total amount of coin transferred between accounts in the epoch. Incl coin transactions and smart wallet transactions.
}

type Stats struct {
	Current    Statistics `json:"current"`
	Cumulative Statistics `json:"cumulative"`
}

type Epoch struct {
	Number     int32  `json:"number" bson:"number"`
	Start      uint32 `json:"start" bson:"start"`
	End        uint32 `json:"end" bson:"end"`
	LayerStart uint32 `json:"layerstart" bson:"layerstart"`
	LayerEnd   uint32 `json:"layerend" bson:"layerend"`
	Layers     uint32 `json:"layers" bson:"layers"`
	Stats      Stats  `json:"stats"`
}

type EpochService interface {
	GetEpoch(ctx context.Context, epochNum int) (*Epoch, error)
	GetEpochs(ctx context.Context, page, perPage int64) (epochs []*Epoch, total int64, err error)
	GetEpochLayers(ctx context.Context, epochNum int, page, perPage int64) (layers []*Layer, total int64, err error)
	GetEpochTransactions(ctx context.Context, epochNum int, page, perPage int64) (txs []*Transaction, total int64, err error)
	GetEpochSmeshers(ctx context.Context, epochNum int, page, perPage int64) (smeshers []*Smesher, total int64, err error)
	GetEpochRewards(ctx context.Context, epochNum int, page, perPage int64) (rewards []*Reward, total int64, err error)
	GetEpochActivations(ctx context.Context, epochNum int, page, perPage int64) (atxs []*Activation, total int64, err error)
}
