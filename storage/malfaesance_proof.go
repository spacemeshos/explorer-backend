package storage

import (
	"context"
	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/go-spacemesh/log"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (s *Storage) SaveMalfeasanceProof(parent context.Context, in *model.MalfeasanceProof) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("malfeasance_proofs").InsertOne(ctx, bson.D{
		{Key: "smesher", Value: in.Smesher},
		{Key: "layer", Value: in.Layer},
		{Key: "type", Value: in.Type},
	})
	if err != nil {
		log.Info("SaveMalfeasanceProof: %v", err)
	}
	return err
}
