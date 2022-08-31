package model

import (
	"context"

	"github.com/spacemeshos/address"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Activation struct {
	Id        string `json:"id" bson:"id"`             //nolint will fix it later.
	Layer     uint32 `json:"layer" bson:"layer"`       // the layer that this activation is part of
	SmesherId string `json:"smesher" bson:"smesher"`   //nolint will fix it later // id of smesher who created the ATX
	Coinbase  string `json:"coinbase" bson:"coinbase"` // coinbase account id
	PrevAtx   string `json:"prevAtx" bson:"prevAtx"`   // previous ATX pointed to
	NumUnits  uint32 `json:"numunits" bson:"numunits"` // number of PoST data commitment units
	Timestamp uint32 `json:"timestamp" bson:"timestamp"`
}

type ActivationService interface {
	GetActivations(ctx context.Context, page, perPage int64) (atxs []*Activation, total int64, err error)
	GetActivation(ctx context.Context, activationID string) (*Activation, error)
}

func NewActivation(atx *pb.Activation, timestamp uint32) *Activation {
	// todo addr cast to string will panic if wrong data in bytes slice.
	// add method validate to address package to check if bytes slice is valid.
	addr := address.GenerateAddress(atx.GetSmesherId().GetId())

	return &Activation{
		Id:        utils.BytesToHex(atx.GetId().GetId()),
		Layer:     atx.GetLayer().GetNumber(),
		SmesherId: addr.String(),
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
