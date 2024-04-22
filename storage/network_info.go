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
	cursor, err := s.db.Collection("networkinfo").Find(ctx, bson.D{{Key: "id", Value: 1}})
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
	_, err := s.db.Collection("networkinfo").UpdateOne(ctx, bson.D{{Key: "id", Value: 1}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "id", Value: 1},
			{Key: "genesisid", Value: in.GenesisId},
			{Key: "genesis", Value: in.GenesisTime},
			{Key: "layers", Value: in.EpochNumLayers},
			{Key: "maxtx", Value: in.MaxTransactionsPerSecond},
			{Key: "duration", Value: in.LayerDuration},
			{Key: "postUnitSize", Value: in.PostUnitSize},
			{Key: "lastlayer", Value: in.LastLayer},
			{Key: "lastlayerts", Value: in.LastLayerTimestamp},
			{Key: "lastapprovedlayer", Value: in.LastApprovedLayer},
			{Key: "lastconfirmedlayer", Value: in.LastConfirmedLayer},
			{Key: "connectedpeers", Value: in.ConnectedPeers},
			{Key: "issynced", Value: in.IsSynced},
			{Key: "syncedlayer", Value: in.SyncedLayer},
			{Key: "toplayer", Value: in.TopLayer},
			{Key: "verifiedlayer", Value: in.VerifiedLayer},
		}},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveOrUpdateNetworkInfo: %v", err)
	}
	return err
}
