package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Block struct {
    Id		string
    Layer	uint32
    Epoch	uint32
    Start	uint32
    End		uint32
    TxsNumber	uint32
    TxsValue	uint64
}

type BlockService interface {
    GetBlock(ctx context.Context, query *bson.D) (*Block, error)
    GetBlocks(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Block, error)
    SaveBlock(ctx context.Context, in *Block) error
}
