package model

import (
	"context"
)

type Block struct {
	Id        string `json:"id" bson:"id"` // nolint will fix it later
	Layer     uint32 `json:"layer" bson:"layer"`
	Epoch     uint32 `json:"epoch" bson:"epoch"`
	Start     uint32 `json:"start" bson:"start"`
	End       uint32 `json:"end" bson:"end"`
	TxsNumber uint32 `json:"txsnumber" bson:"txsnumber"`
	TxsValue  uint64 `json:"txsvalue" bson:"txsvalue"`
}

type BlockService interface {
	GetBlock(ctx context.Context, blockID string) (*Block, error)
}
