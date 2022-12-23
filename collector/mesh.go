package collector

import (
	"context"
	"github.com/spacemeshos/explorer-backend/utils"
	"io"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/log"
)

func (c *Collector) getNetworkInfo() error {
	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	genesisTime, err := c.meshClient.GenesisTime(ctx, &pb.GenesisTimeRequest{})
	if err != nil {
		log.Error("cannot get GenesisTime: %v", err)
		return err
	}

	genesisId, err := c.meshClient.GenesisID(ctx, &pb.GenesisIDRequest{})
	if err != nil {
		log.Error("cannot get NetId: %v", err)
	}

	epochNumLayers, err := c.meshClient.EpochNumLayers(ctx, &pb.EpochNumLayersRequest{})
	if err != nil {
		log.Error("cannot get EpochNumLayers: %v", err)
		return err
	}

	maxTransactionsPerSecond, err := c.meshClient.MaxTransactionsPerSecond(ctx, &pb.MaxTransactionsPerSecondRequest{})
	if err != nil {
		log.Error("cannot get MaxTransactionsPerSecond: %v", err)
		return err
	}

	layerDuration, err := c.meshClient.LayerDuration(ctx, &pb.LayerDurationRequest{})
	if err != nil {
		log.Error("cannot get LayerDuration: %v", err)
		return err
	}

	accounts, err := c.debugClient.Accounts(ctx, &empty.Empty{})
	if err != nil {
		log.Error("cannot get accounts: %v", err)
		return err
	}

	res, err := c.smesherClient.PostConfig(ctx, &empty.Empty{})
	if err != nil {
		log.Error("cannot get POST config: %v", err)
		return err
	}

	c.listener.OnNetworkInfo(
		utils.BytesToHex(genesisId.GetGenesisId()),
		genesisTime.GetUnixtime().GetValue(),
		epochNumLayers.GetNumlayers().GetValue(),
		maxTransactionsPerSecond.GetMaxTxsPerSecond().GetValue(),
		layerDuration.GetDuration().GetValue(),
		(uint64(res.BitsPerLabel)*uint64(res.LabelsPerUnit))/8,
	)

	for _, account := range accounts.GetAccountWrapper() {
		c.listener.OnAccount(account)
	}

	return nil
}

func (c *Collector) layersPump() error {
	var req pb.LayerStreamRequest

	log.Info("Start mesh layer pump")
	defer func() {
		c.notify <- -streamType_mesh_Layer
		log.Info("Stop mesh layer pump")
	}()

	c.notify <- +streamType_mesh_Layer

	stream, err := c.meshClient.LayerStream(context.Background(), &req)
	if err != nil {
		log.Error("cannot get layer stream: %v", err)
		return err
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			log.Error("cannot receive layer: %v", err)
			return err
		}
		layer := response.GetLayer()
		c.listener.OnLayer(layer)
	}
}

func (c *Collector) syncMissingLayers() error {
	status, err := c.nodeClient.Status(context.Background(), &pb.StatusRequest{})
	if err != nil {
		log.Error("cannot receive node status: %v", err)
		return err
	}
	syncedLayerNum := status.Status.SyncedLayer.Number
	lastLayer := c.listener.GetLastLayer(context.TODO())

	if syncedLayerNum == lastLayer {
		return nil
	}

	layers, err := c.meshClient.LayersQuery(context.Background(), &pb.LayersQueryRequest{
		StartLayer: &pb.LayerNumber{Number: lastLayer + 1},
		EndLayer:   &pb.LayerNumber{Number: syncedLayerNum},
	})
	if err != nil {
		return err
	}

	for _, layer := range layers.GetLayer() {
		log.Info("syncing missing layer: %d", layer.Number.Number)
		c.listener.OnLayer(layer)
	}

	return nil
}
