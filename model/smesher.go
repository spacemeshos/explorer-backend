package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Geo struct {
    Name	string `json:"name"`
    Coordinates	[2]float64 `json:"coordinates"`
}

type Smesher struct {
	Id             string
	Geo            Geo
	CommitmentSize uint64 `json:"cSize"`
	Coinbase       string
	AtxCount       uint32
	Timestamp      uint32
}

type SmesherService interface {
    GetSmesher(ctx context.Context, query *bson.D) (*Smesher, error)
    GetSmeshers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Smesher, error)
    SaveSmesher(ctx context.Context, in *Smesher) error
}
