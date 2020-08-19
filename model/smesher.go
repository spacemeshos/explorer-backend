package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
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
