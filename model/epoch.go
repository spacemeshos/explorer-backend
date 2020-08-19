package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
)

type Statistics struct {
    Capacity		uint64	// Average tx/s rate over capacity considering all layers in the current epoch.
    Decentral		uint64	// Distribution of storage between all active smeshers.
    Smeshers		uint64	// Number of active smeshers in the current epoch.
    Transactions	uint64	// Total number of transactions processed by the state transition function.
    Accounts		uint64	// Total number of on-mesh accounts with a non-zero coin balance as of the current epoch.
    Circulation		uint64	// Total number of Smesh coins in circulation. This is the total balances of all on-mesh accounts.
    Rewards		uint64	// Total amount of Smesh minted as mining rewards as of the last known reward distribution event.
    Security		uint64	// Total amount of storage committed to the network based on the ATXs in the previous epoch.
}

type Stats struct {
    Current	Statistics
    Cumulative	Statistics
}

type Epoch struct {
    Number	int32
    Stats	Stats
}

type EpochService interface {
    GetEpoch(ctx context.Context, query *bson.D) (*Epoch, error)
    GetEpochs(ctx context.Context, query *bson.D) ([]*Epoch, error)
    SaveEpoch(ctx context.Context, in *Epoch) error
}
