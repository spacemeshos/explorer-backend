package model

import (
	"context"
	"github.com/spacemeshos/go-spacemesh/common/types"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Activation struct {
	Id                string           `json:"id" bson:"id"`             //nolint will fix it later.
	SmesherId         string           `json:"smesher" bson:"smesher"`   //nolint will fix it later // id of smesher who created the ATX
	Coinbase          string           `json:"coinbase" bson:"coinbase"` // coinbase account id
	PrevAtx           string           `json:"prevAtx" bson:"prevAtx"`   // previous ATX pointed to
	NumUnits          uint32           `json:"numunits" bson:"numunits"` // number of PoST data commitment units
	CommitmentSize    uint64           `json:"commitmentSize" bson:"commitmentSize"`
	PublishEpoch      uint32           `json:"publishEpoch" bson:"publishEpoch"`
	TargetEpoch       uint32           `json:"targetEpoch" bson:"targetEpoch"`
	TickCount         uint64           `json:"tickCount" bson:"tickCount"`
	Weight            uint64           `json:"weight" bson:"weight"`
	EffectiveNumUnits uint32           `json:"effectiveNumUnits" bson:"effectiveNumUnits"`
	Received          map[string]int64 `json:"received" bson:"cReceived"`
}

type ActivationService interface {
	GetActivations(ctx context.Context, page, perPage int64) (atxs []*Activation, total int64, err error)
	GetActivation(ctx context.Context, activationID string) (*Activation, error)
}

func NewActivation(atx *types.VerifiedActivationTx, collectorName string) *Activation {
	return &Activation{
		Id:                utils.BytesToHex(atx.ID().Bytes()),
		PublishEpoch:      atx.PublishEpoch.Uint32(),
		TargetEpoch:       atx.PublishEpoch.Uint32() + 1,
		SmesherId:         utils.BytesToHex(atx.SmesherID.Bytes()),
		Coinbase:          atx.Coinbase.String(),
		PrevAtx:           utils.BytesToHex(atx.PrevATXID.Bytes()),
		NumUnits:          atx.NumUnits,
		TickCount:         atx.TickCount(),
		Weight:            atx.GetWeight(),
		EffectiveNumUnits: atx.EffectiveNumUnits(),
		Received: map[string]int64{
			collectorName: atx.Received().UnixNano(),
		},
	}
}

func (atx *Activation) GetSmesher(unitSize uint64, collectorName string) *Smesher {
	return &Smesher{
		Id:             atx.SmesherId,
		Coinbase:       atx.Coinbase,
		Timestamp:      uint64(atx.Received[collectorName]),
		CommitmentSize: uint64(atx.NumUnits) * unitSize,
	}
}
