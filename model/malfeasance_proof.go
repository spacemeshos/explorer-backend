package model

import (
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/explorer-backend/utils"
)

type MalfeasanceProof struct {
	Smesher string `json:"smesher" bson:"smesher"`
	Layer   uint32 `json:"layer" bson:"layer"`
	Type    string `json:"type" bson:"type"`
}

func NewMalfeasanceProof(in *pb.MalfeasanceProof) *MalfeasanceProof {
	return &MalfeasanceProof{
		Smesher: utils.BytesToHex(in.GetSmesherId().GetId()),
		Layer:   in.Layer.GetNumber(),
		Type:    in.Kind.String(),
	}
}