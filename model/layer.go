package model

import (
	"context"
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Layer struct {
	Number       uint32 `json:"number" bson:"number"`
	Status       int    `json:"status" bson:"status"`
	Txs          uint32 `json:"txs" bson:"txs"`
	Start        uint32 `json:"start" bson:"start"`
	End          uint32 `json:"end" bson:"end"`
	TxsAmount    uint64 `json:"txsamount" bson:"txsamount"`
	Rewards      uint64 `json:"rewards" bson:"rewards"`
	Epoch        uint32 `json:"epoch" bson:"epoch"`
	Hash         string `json:"hash" bson:"hash"`
	BlocksNumber uint32 `json:"blocksnumber" bson:"blocksnumber"`
}

type LayerService interface {
	GetLayer(ctx context.Context, layerNum int) (*Layer, error)
	//GetLayerByHash(ctx context.Context, layerHash string) (*Layer, error)
	GetLayers(ctx context.Context, page, perPage int64) (layers []*Layer, total int64, err error)
	GetLayerTransactions(ctx context.Context, layerNum int, pageNum, pageSize int64) (txs []*Transaction, total int64, err error)
	GetLayerSmeshers(ctx context.Context, layerNum int, pageNum, pageSize int64) (smeshers []*Smesher, total int64, err error)
	GetLayerRewards(ctx context.Context, layerNum int, pageNum, pageSize int64) (rewards []*Reward, total int64, err error)
	GetLayerActivations(ctx context.Context, layerNum int, pageNum, pageSize int64) (atxs []*Activation, total int64, err error)
	GetLayerBlocks(ctx context.Context, layerNum int, pageNum, pageSize int64) (blocks []*Block, total int64, err error)
}

func NewLayer(in *pb.Layer, networkInfo *NetworkInfo) (*Layer, []*Block, []*Activation, map[string]*Transaction) {
	pbBlocks := in.GetBlocks()
	pbAtxs := in.GetActivations()
	layer := &Layer{
		Number:       in.Number.Number,
		Status:       int(in.GetStatus()),
		Epoch:        in.Number.Number / networkInfo.EpochNumLayers,
		BlocksNumber: uint32(len(pbBlocks)),
		Hash:         utils.BytesToHex(in.Hash),
	}
	if layer.Number == 0 {
		layer.Start = networkInfo.GenesisTime
	} else {
		layer.Start = networkInfo.GenesisTime + layer.Number*networkInfo.LayerDuration
	}
	layer.End = layer.Start + networkInfo.LayerDuration - 1

	blocks := make([]*Block, len(pbBlocks))
	atxs := make([]*Activation, len(pbAtxs))
	txs := make(map[string]*Transaction)

	for i, b := range pbBlocks {
		blocks[i] = &Block{
			Id:        utils.NBytesToHex(b.GetId(), 20),
			Layer:     layer.Number,
			Epoch:     layer.Epoch,
			Start:     layer.Start,
			End:       layer.End,
			TxsNumber: uint32(len(b.Transactions)),
		}
		for j, t := range b.Transactions {
			tx, err := NewTransaction(t, layer.Number, blocks[i].Id, layer.Start, uint32(j))
			if err != nil {
				log.Err(fmt.Errorf("cannot create transaction: %v", err))
				continue
			}
			txs[tx.Id] = tx
			blocks[i].TxsValue += tx.Amount
		}
	}

	layer.Txs = uint32(len(txs))
	for _, tx := range txs {
		layer.TxsAmount += tx.Amount
	}

	return layer, blocks, atxs, txs
}

func IsApprovedLayer(l *pb.Layer) bool {
	return l.GetStatus() >= pb.Layer_LAYER_STATUS_APPROVED
}

func IsConfirmedLayer(l *pb.Layer) bool {
	return l.GetStatus() == pb.Layer_LAYER_STATUS_CONFIRMED
}
