package model

import (
    "go.mongodb.org/mongo-driver/bson"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Geo struct {
    Name	string `json:"name"`
    Coordinates	[2]float64 `json:"coordinates"`
}

type Smesher struct {
    Id			string
    Geo			Geo
    CommitmentSize	uint64	// commitment size in bytes
}

type SmesherService interface {
    GetSmesher(ctx context.Context, query *bson.D) (*Smesher, error)
    GetSmeshers(ctx context.Context, query *bson.D) ([]*Smesher, error)
    SaveSmesher(ctx context.Context, in *Smesher) error
}
