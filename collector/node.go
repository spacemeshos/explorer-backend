package collector

import (
	"context"
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"io"

	"github.com/spacemeshos/go-spacemesh/log"
)

func (c *Collector) syncStatusPump() error {
	req := pb.StatusStreamRequest{}

	log.Info("Start node sync status pump")
	defer func() {
		c.notify <- -streamType_node_SyncStatus
		log.Info("Stop node sync status pump")
	}()

	c.notify <- +streamType_node_SyncStatus

	stream, err := c.nodeClient.StatusStream(context.Background(), &req)
	if err != nil {
		log.Err(fmt.Errorf("cannot get sync status stream: %v", err))
		return err
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			log.Info("syncStatusPump: EOF")
			return err
		}
		if err != nil {
			log.Err(fmt.Errorf("cannot receive sync status: %v", err))
			return err
		}

		status := res.GetStatus()
		log.Info("Node sync status: %v", status)

		lastLayer := c.listener.GetLastLayer(context.TODO())
		if lastLayer != status.GetVerifiedLayer().GetNumber() {
			for i := lastLayer + 1; i <= status.GetVerifiedLayer().GetNumber(); i++ {
				err := c.syncLayer(types.LayerID(i))
				if err != nil {
					fmt.Errorf("syncLayer error: %v", err)
				}

				err = c.syncNotProcessedTxs()
				if err != nil {
					fmt.Errorf("syncNotProcessedTxs error: %v", err)
				}

				if c.atxSyncFlag {
					err = c.syncActivations()
					if err != nil {
						fmt.Errorf("syncActivations error: %v", err)
					}
				}

				err = c.createFutureEpoch()
				if err != nil {
					fmt.Errorf("createFutureEpoch error: %v", err)
				}
			}
		}

		c.listener.OnNodeStatus(
			status.GetConnectedPeers(),
			status.GetIsSynced(),
			status.GetSyncedLayer().GetNumber(),
			status.GetTopLayer().GetNumber(),
			status.GetVerifiedLayer().GetNumber(),
		)
	}
}
