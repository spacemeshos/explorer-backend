package storage

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) GetNetworkInfo(parent context.Context) (*model.NetworkInfo, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("networkinfo").Find(ctx, bson.D{{"id", 1}})
	if err != nil {
		log.Info("GetNetworkInfo: %v", err)
		return nil, err
	}
	if !cursor.Next(ctx) {
		log.Info("GetNetworkInfo: Empty result")
		return nil, errors.New("Empty result")
	}
	doc := cursor.Current
	info := &model.NetworkInfo{
		GenesisId:                utils.GetAsString(doc.Lookup("genesisid")),
		GenesisTime:              utils.GetAsUInt32(doc.Lookup("genesis")),
		EpochNumLayers:           utils.GetAsUInt32(doc.Lookup("layers")),
		MaxTransactionsPerSecond: utils.GetAsUInt32(doc.Lookup("maxtx")),
		LayerDuration:            utils.GetAsUInt32(doc.Lookup("duration")),
		LastLayer:                utils.GetAsUInt32(doc.Lookup("lastlayer")),
		LastLayerTimestamp:       utils.GetAsUInt32(doc.Lookup("lastlayerts")),
		LastApprovedLayer:        utils.GetAsUInt32(doc.Lookup("lastapprovedlayer")),
		LastConfirmedLayer:       utils.GetAsUInt32(doc.Lookup("lastconfirmedlayer")),
		ConnectedPeers:           utils.GetAsUInt64(doc.Lookup("connectedpeers")),
		IsSynced:                 utils.GetAsBool(doc.Lookup("issynced")),
		SyncedLayer:              utils.GetAsUInt32(doc.Lookup("syncedlayer")),
		TopLayer:                 utils.GetAsUInt32(doc.Lookup("toplayer")),
		VerifiedLayer:            utils.GetAsUInt32(doc.Lookup("verifiedlayer")),
	}
	return info, nil
}

func (s *Storage) SaveOrUpdateNetworkInfo(parent context.Context, in *model.NetworkInfo) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("networkinfo").UpdateOne(ctx, bson.D{{"id", 1}}, bson.D{
		{"$set", bson.D{
			{"id", 1},
			{"genesisid", in.GenesisId},
			{"genesis", in.GenesisTime},
			{"layers", in.EpochNumLayers},
			{"maxtx", in.MaxTransactionsPerSecond},
			{"duration", in.LayerDuration},
			{"lastlayer", in.LastLayer},
			{"lastlayerts", in.LastLayerTimestamp},
			{"lastapprovedlayer", in.LastApprovedLayer},
			{"lastconfirmedlayer", in.LastConfirmedLayer},
			{"connectedpeers", in.ConnectedPeers},
			{"issynced", in.IsSynced},
			{"syncedlayer", in.SyncedLayer},
			{"toplayer", in.TopLayer},
			{"verifiedlayer", in.VerifiedLayer},
		}},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveOrUpdateNetworkInfo: %v", err)
	}
	return err
}
