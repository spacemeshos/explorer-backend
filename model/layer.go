package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Layer struct {
    Number	uint32
    Status	int
    Txs		uint32
    Start	uint32
    End		uint32
    TxsAmount	uint64
    AtxCSize	uint64
    Rewards	uint64
}

type LayerService interface {
    GetLayer(ctx context.Context, query *bson.D) (*Layer, error)
    GetLayers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Layer, error)
    SaveLayer(ctx context.Context, in *Layer) error
}

func NewLayer(l *pb.Layer, genesisTime uint32, layerDuration uint32) (*Layer, []*Block, []*Activation, map[string]*Transaction) {
    pbBlocks := l.GetBlocks()
    pbAtxs := l.GetActivations()
    layer := &Layer{
        Number: l.GetNumber().GetNumber(),
        Status: int(l.GetStatus()),
    }
    layer.Start = genesisTime + layer.Number * layerDuration
    layer.End = layer.Start + layerDuration - 1

    blocks := make([]*Block, len(pbBlocks))
    atxs := make([]*Activation, len(pbAtxs))
    txs := make(map[string]*Transaction)

    for i, b := range pbBlocks {
        blocks[i] = &Block{
            Id: utils.BytesToHex(b.GetId()),
            Layer: layer.Number,
        }
        for j, t := range b.GetTransactions() {
            tx := NewTransaction(t, layer.Number, blocks[i].Id, layer.Start, uint32(j))
            txs[tx.Id] = tx
        }
    }

    layer.Txs = uint32(len(txs))
    for _, tx := range txs {
        layer.TxsAmount += tx.Amount
    }

    for i, a := range pbAtxs {
        atxs[i] = NewActivation(a)
        layer.AtxCSize += atxs[i].CommitmentSize
    }

    return layer, blocks, atxs, txs
}

func IsApprovedLayer(l *pb.Layer) bool {
    return l.GetStatus() >= pb.Layer_LAYER_STATUS_APPROVED
}

func IsConfirmedLayer(l *pb.Layer) bool {
    return l.GetStatus() == pb.Layer_LAYER_STATUS_CONFIRMED
}
