package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Statistics struct {
    Capacity		int64	// Average tx/s rate over capacity considering all layers in the current epoch.
    Decentral		int64	// Distribution of storage between all active smeshers.
    Smeshers		int64	// Number of active smeshers in the current epoch.
    Transactions	int64	// Total number of transactions processed by the state transition function.
    Accounts		int64	// Total number of on-mesh accounts with a non-zero coin balance as of the current epoch.
    Circulation		int64	// Total number of Smesh coins in circulation. This is the total balances of all on-mesh accounts.
    Rewards		int64	// Total amount of Smesh minted as mining rewards as of the last known reward distribution event.
    Security		int64	// Total amount of storage committed to the network based on the ATXs in the previous epoch.
    TxsAmount		int64	// Total amount of coin transferred between accounts in the epoch. Incl coin transactions and smart wallet transactions.
}

type Stats struct {
    Current	Statistics
    Cumulative	Statistics
}

type Epoch struct {
    Number	int32
    Start	uint32
    End		uint32
    LayerStart	uint32
    LayerEnd	uint32
    Layers	uint32
    Stats	Stats
}

type EpochService interface {
    GetEpoch(ctx context.Context, query *bson.D) (*Epoch, error)
    GetEpochs(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Epoch, error)
    SaveEpoch(ctx context.Context, in *Epoch) error
}
