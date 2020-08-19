package model

import (
    "go.mongodb.org/mongo-driver/bson"
)

type Block struct {
    Id		string
    Layer	uint64
}

type BlockService interface {
    GetBlock(ctx context.Context, query *bson.D) (*Block, error)
    GetBlocks(ctx context.Context, query *bson.D) ([]*Block, error)
    SaveBlock(ctx context.Context, in *Block) error
}
