package model

import (
    "go.mongodb.org/mongo-driver/bson"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Layer struct {
    Number	uint64
    Status	int
}

type LayerService interface {
    GetLayer(ctx context.Context, query *bson.D) (*Layer, error)
    GetLayers(ctx context.Context, query *bson.D) ([]*Layer, error)
    SaveLayer(ctx context.Context, in *Layer) error
}

func NewLayer(l *pb.Layer) (*Layer, []*Block, []*Activation, make(map[string]*Transaction) {
    pbBlocks := l.GetBlocks()
    pbAtxs := l.GetActivations()
    layer := &Layer{
        Number: LayerID(l.GetNumber().GetNumber()),
        Status: l.GetStatus(),
    }

    blocks := make([]*Block, len(blocks)),
    atxs := make([]*Activation, len(atxs)),
    txs := make(map[string]Transaction),

    for i, b := range pbBlocks {
        blocks[i] := &Block{
            Id: utils.ToHex(b.GetId()),
            Layer: layer.Number,
        }
        for _, t := range b.GetTransactions() {
            tx := NewTransaction(t, layer.Number, blocks[i].Id)
            txs[tx.Id] = tx
        }
    }

    for i, a := range pbAtxs {
        atxs[i] = NewActivation(a)
    }

    return layer, blocks, atxs, txs
}

func IsApprovedLayer(l *pb.Layer) bool {
    return l.GetStatus() >= pb.Layer_LAYER_STATUS_APPROVED
}

func IsConfirmedLayer(l *pb.Layer) bool {
    return l.GetStatus() == pb.Layer_LAYER_STATUS_CONFIRMED
}
