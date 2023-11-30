package model

import (
	"context"
)

type Geo struct {
	Name        string     `json:"name"`
	Coordinates [2]float64 `json:"coordinates"`
}

type Smesher struct {
	Id             string             `json:"id" bson:"id"` //nolint will fix it later.
	CommitmentSize uint64             `json:"cSize" bson:"cSize"`
	Coinbase       string             `json:"coinbase" bson:"coinbase"`
	AtxCount       uint32             `json:"atxcount" bson:"atxcount"`
	Timestamp      uint32             `json:"timestamp" bson:"timestamp"`
	Name           string             `json:"name" bson:"name"`
	Lat            float64            `json:"lat" bson:"lat"`
	Lon            float64            `json:"lon" bson:"lon"`
	Rewards        int64              `json:"rewards" bson:"-"`
	Proofs         []MalfeasanceProof `json:"proofs,omitempty" bson:"proofs,omitempty"`
}

type SmesherService interface {
	GetSmesher(ctx context.Context, smesherID string) (*Smesher, error)
	GetSmeshers(ctx context.Context, page, perPage int64) (smeshers []*Smesher, total int64, err error)
	GetSmesherActivations(ctx context.Context, smesherID string, page, perPage int64) (atxs []*Activation, total int64, err error)
	GetSmesherRewards(ctx context.Context, smesherID string, page, perPage int64) (rewards []*Reward, total int64, err error)
	CountSmesherRewards(ctx context.Context, smesherID string) (total, count int64, err error)
}
