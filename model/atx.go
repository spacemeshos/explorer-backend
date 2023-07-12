package model

import (
	"context"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Activation struct {
	Id             string `json:"id" bson:"id"`             //nolint will fix it later.
	Layer          uint32 `json:"layer" bson:"layer"`       // the layer that this activation is part of
	SmesherId      string `json:"smesher" bson:"smesher"`   //nolint will fix it later // id of smesher who created the ATX
	Coinbase       string `json:"coinbase" bson:"coinbase"` // coinbase account id
	PrevAtx        string `json:"prevAtx" bson:"prevAtx"`   // previous ATX pointed to
	NumUnits       uint32 `json:"numunits" bson:"numunits"` // number of PoST data commitment units
	CommitmentSize uint64 `json:"commitmentSize" bson:"commitmentSize"`
	Timestamp      uint32 `json:"timestamp" bson:"timestamp"`
}

type ActivationService interface {
	GetActivations(ctx context.Context, page, perPage int64) (atxs []*Activation, total int64, err error)
	GetActivation(ctx context.Context, activationID string) (*Activation, error)
}

func NewActivation(atx *pb.Activation, timestamp uint32) *Activation {
	return &Activation{
		Id:        utils.BytesToHex(atx.GetId().GetId()),
		Layer:     atx.GetLayer().GetNumber(),
		SmesherId: utils.BytesToHex(atx.GetSmesherId().GetId()),
		Coinbase:  atx.GetCoinbase().GetAddress(),
		PrevAtx:   utils.BytesToHex(atx.GetPrevAtx().GetId()),
		NumUnits:  atx.GetNumUnits(),
		Timestamp: timestamp,
	}
}

func (atx *Activation) GetSmesher(unitSize uint64) *Smesher {
	return &Smesher{
		Id:             atx.SmesherId,
		CommitmentSize: uint64(atx.NumUnits) * unitSize,
	}
}
