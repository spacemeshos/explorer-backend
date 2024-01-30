package collector

import (
	"context"
	"fmt"
	"github.com/spacemeshos/explorer-backend/utils"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"go.mongodb.org/mongo-driver/bson"
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
		log.Err(fmt.Errorf("cannot get GenesisTime: %v", err))
		return err
	}

	genesisId, err := c.meshClient.GenesisID(ctx, &pb.GenesisIDRequest{})
	if err != nil {
		log.Err(fmt.Errorf("cannot get NetId: %v", err))
	}

	epochNumLayers, err := c.meshClient.EpochNumLayers(ctx, &pb.EpochNumLayersRequest{})
	if err != nil {
		log.Err(fmt.Errorf("cannot get EpochNumLayers: %v", err))
		return err
	}

	maxTransactionsPerSecond, err := c.meshClient.MaxTransactionsPerSecond(ctx, &pb.MaxTransactionsPerSecondRequest{})
	if err != nil {
		log.Err(fmt.Errorf("cannot get MaxTransactionsPerSecond: %v", err))
		return err
	}

	layerDuration, err := c.meshClient.LayerDuration(ctx, &pb.LayerDurationRequest{})
	if err != nil {
		log.Err(fmt.Errorf("cannot get LayerDuration: %v", err))
		return err
	}

	res, err := c.smesherClient.PostConfig(ctx, &empty.Empty{})
	if err != nil {
		log.Err(fmt.Errorf("cannot get POST config: %v", err))
		return err
	}

	c.listener.OnNetworkInfo(
		utils.BytesToHex(genesisId.GetGenesisId()),
		genesisTime.GetUnixtime().GetValue(),
		epochNumLayers.GetNumlayers().GetNumber(),
		maxTransactionsPerSecond.GetMaxTxsPerSecond().GetValue(),
		layerDuration.GetDuration().GetValue(),
		(uint64(res.BitsPerLabel)*uint64(res.LabelsPerUnit))/8,
	)

	return nil
}

func (c *Collector) syncMissingLayers() error {
	status, err := c.nodeClient.Status(context.Background(), &pb.StatusRequest{})
	if err != nil {
		log.Err(fmt.Errorf("cannot receive node status: %v", err))
		return err
	}
	syncedLayerNum := status.Status.VerifiedLayer.Number
	lastLayer := c.listener.GetLastLayer(context.TODO())

	if syncedLayerNum == lastLayer {
		return nil
	}

	log.Info("Syncing missing layers %d...%d", lastLayer+1, syncedLayerNum)

	for i := lastLayer + 1; i <= syncedLayerNum; i++ {
		err := c.syncLayer(types.LayerID(i))
		if err != nil {
			fmt.Errorf("syncMissingLayers error: %v", err)
		}
	}

	log.Info("Waiting for layers queue to be empty")
	for {
		layersInQueue := c.listener.LayersInQueue()
		if layersInQueue > 0 {
			log.Info("%d layers in queue. Waiting", layersInQueue)
			time.Sleep(15 * time.Second)
		} else {
			break
		}
	}

	return nil
}

func (c *Collector) malfeasancePump() error {
	var req = pb.MalfeasanceStreamRequest{}

	log.Info("Start mesh malfeasance pump")
	defer func() {
		c.notify <- -streamType_mesh_Malfeasance
		log.Info("Stop mesh malfeasance pump")
	}()

	c.notify <- +streamType_mesh_Malfeasance

	stream, err := c.meshClient.MalfeasanceStream(context.Background(), &req)
	if err != nil {
		log.Err(fmt.Errorf("cannot get malfeasance stream: %v", err))
		return err
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			log.Err(fmt.Errorf("cannot receive malfeasance proof: %v", err))
			return err
		}
		proof := response.GetProof()
		c.listener.OnMalfeasanceProof(proof)
	}
}

func (c *Collector) syncLayer(lid types.LayerID) error {
	layer, err := c.dbClient.GetLayer(c.db, lid, c.listener.GetEpochNumLayers())
	if err != nil {
		return err
	}

	if c.listener.IsLayerInQueue(layer) {
		log.Info("layer %d is already in queue", layer.Number.Number)
		return nil
	}

	if lastLayer := c.listener.GetLastLayer(context.TODO()); lastLayer >= layer.Number.Number {
		log.Info("layer %d is already in database", layer.Number.Number)
		return nil
	}

	log.Info("syncing layer: %d", layer.Number.Number)
	c.listener.OnLayer(layer)

	log.Info("syncing accounts for layer: %d", layer.Number.Number)
	accounts, err := c.dbClient.AccountsSnapshot(c.db, lid)
	if err != nil {
		fmt.Errorf("%v\n", err)
	}
	c.listener.OnAccounts(accounts)

	log.Info("syncing rewards for layer: %d", layer.Number.Number)
	rewards, err := c.dbClient.GetLayerRewards(c.db, lid)
	if err != nil {
		fmt.Errorf("%v\n", err)
	}

	for _, reward := range rewards {
		r := &pb.Reward{
			Layer:       &pb.LayerNumber{Number: reward.Layer.Uint32()},
			Total:       &pb.Amount{Value: reward.TotalReward},
			LayerReward: &pb.Amount{Value: reward.LayerReward},
			Coinbase:    &pb.AccountId{Address: reward.Coinbase.String()},
			Smesher:     &pb.SmesherId{Id: reward.SmesherID.Bytes()},
		}
		c.listener.OnReward(r)
	}

	c.listener.UpdateEpochStats(layer.Number.Number)

	return nil
}

func (c *Collector) syncNotProcessedTxs() error {
	txs, err := c.listener.GetTransactions(context.TODO(), &bson.D{{Key: "state", Value: 0}})
	if err != nil {
		return err
	}

	for _, tx := range txs {
		txId, err := utils.StringToBytes(tx.Id)
		if err != nil {
			return err
		}

		state, err := c.transactionsClient.TransactionsState(context.TODO(), &pb.TransactionsStateRequest{
			TransactionId:       []*pb.TransactionId{{Id: txId}},
			IncludeTransactions: false,
		})
		if err != nil {
			return err
		}

		txState := state.TransactionsState[0]

		if txState != nil {
			err := c.listener.UpdateTransactionState(context.TODO(), tx.Id, int32(txState.State))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Collector) syncAllRewards() error {
	rewards, err := c.dbClient.GetAllRewards(c.db)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}

	for _, reward := range rewards {
		r := &pb.Reward{
			Layer:       &pb.LayerNumber{Number: reward.Layer.Uint32()},
			Total:       &pb.Amount{Value: reward.TotalReward},
			LayerReward: &pb.Amount{Value: reward.LayerReward},
			Coinbase:    &pb.AccountId{Address: reward.Coinbase.String()},
			Smesher:     &pb.SmesherId{Id: reward.SmesherID.Bytes()},
		}
		c.listener.OnReward(r)
	}

	return nil
}
