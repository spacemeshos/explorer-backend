package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
)

type Block struct {
    Id		string
    Layer	uint32
}

type BlockService interface {
    GetBlock(ctx context.Context, query *bson.D) (*Block, error)
    GetBlocks(ctx context.Context, query *bson.D) ([]*Block, error)
    SaveBlock(ctx context.Context, in *Block) error
}
