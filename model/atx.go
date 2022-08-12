package model

import (
	"context"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Activation struct {
	Id        string //nolint will fix it later.
	Layer     uint32 // the layer that this activation is part of
	SmesherId string `json:"smesher"` //nolint will fix it later // id of smesher who created the ATX
	Coinbase  string // coinbase account id
	PrevAtx   string // previous ATX pointed to
	NumUnits  uint32 // number of PoST data commitment units
	Timestamp uint32
}

type ActivationService interface {
	GetActivation(ctx context.Context, query *bson.D) (*Activation, error)
	GetActivations(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Activation, error)
	SaveActivation(ctx context.Context, in *Activation) error
}

func NewActivation(atx *pb.Activation, timestamp uint32) *Activation {
	return &Activation{
		Id:        utils.BytesToHex(atx.GetId().GetId()),
		Layer:     atx.GetLayer().GetNumber(),
		SmesherId: utils.BytesToHex(atx.GetSmesherId().GetId()),
		Coinbase:  utils.BytesToAddressString(atx.GetCoinbase().GetAddress()),
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
