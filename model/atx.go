package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Activation struct {
    Id			string
    Layer		uint32	// the layer that this activation is part of
    SmesherId		string	// id of smesher who created the ATX
    Coinbase		string	// coinbase account id
    PrevAtx		string	// previous ATX pointed to
    CommitmentSize	uint64	// commitment size in bytes
}

type ActivationService interface {
    GetActivation(ctx context.Context, query *bson.D) (*Activation, error)
    GetActivations(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Activation, error)
    SaveActivation(ctx context.Context, in *Activation) error
}

func NewActivation(atx *pb.Activation) *Activation {
    return &Activation{
        Id: utils.BytesToHex(atx.GetId().GetId()),
        Layer: atx.GetLayer().GetNumber(),
        SmesherId: utils.BytesToHex(atx.GetSmesherId().GetId()),
        Coinbase: utils.BytesToAddressString(atx.GetCoinbase().GetAddress()),
        PrevAtx: utils.BytesToHex(atx.GetPrevAtx().GetId()),
        CommitmentSize: atx.GetCommitmentSize(),
    }
}

func (atx *Activation) GetSmesher() *Smesher {
    return &Smesher{
        Id: atx.SmesherId,
        CommitmentSize: atx.CommitmentSize,
    }
}
