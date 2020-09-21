package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Layer struct {
    Number		uint32
    Status		int
    Txs			uint32
    Start		uint32
    End			uint32
    TxsAmount		uint64
    AtxCSize		uint64
    Rewards		uint64
    Epoch		uint32
    Smeshers		uint32
    Hash		string
    BlocksNumber	uint32
}

type LayerService interface {
    GetLayer(ctx context.Context, query *bson.D) (*Layer, error)
    GetLayers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Layer, error)
    SaveLayer(ctx context.Context, in *Layer) error
}

func NewLayer(in *pb.Layer, networkInfo *NetworkInfo) (*Layer, []*Block, []*Activation, map[string]*Transaction) {
    pbBlocks := in.GetBlocks()
    pbAtxs := in.GetActivations()
    layer := &Layer{
        Number: in.Number.Number,
        Status: int(in.GetStatus()),
        Epoch: in.Number.Number / networkInfo.EpochNumLayers,
        BlocksNumber: uint32(len(pbBlocks)),
        Hash: utils.BytesToHex(in.Hash),
    }
    layer.Start = networkInfo.GenesisTime + layer.Number * networkInfo.LayerDuration
    layer.End = layer.Start + networkInfo.LayerDuration - 1

    blocks := make([]*Block, len(pbBlocks))
    atxs := make([]*Activation, len(pbAtxs))
    txs := make(map[string]*Transaction)

    for i, b := range pbBlocks {
        blocks[i] = &Block{
            Id: utils.BytesToHex(b.GetId()),
            Layer: layer.Number,
            Epoch: layer.Epoch,
            Start: layer.Start,
            End: layer.End,
            TxsNumber: uint32(len(b.Transactions)),
        }
        for j, t := range b.Transactions {
            tx := NewTransaction(t, layer.Number, blocks[i].Id, layer.Start, uint32(j))
            txs[tx.Id] = tx
            blocks[i].TxsValue += tx.Amount
        }
    }

    layer.Txs = uint32(len(txs))
    for _, tx := range txs {
        layer.TxsAmount += tx.Amount
    }

    smeshers := make(map[string]bool)
    for i, a := range pbAtxs {
        atxs[i] = NewActivation(a)
        layer.AtxCSize += atxs[i].CommitmentSize
        smeshers[atxs[i].SmesherId] = true
    }

    layer.Smeshers = uint32(len(smeshers))

    return layer, blocks, atxs, txs
}

func IsApprovedLayer(l *pb.Layer) bool {
    return l.GetStatus() >= pb.Layer_LAYER_STATUS_APPROVED
}

func IsConfirmedLayer(l *pb.Layer) bool {
    return l.GetStatus() == pb.Layer_LAYER_STATUS_CONFIRMED
}
