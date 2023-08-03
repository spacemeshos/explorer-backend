package collector

import (
	"context"
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
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

		c.listener.OnNodeStatus(
			status.GetConnectedPeers(),
			status.GetIsSynced(),
			status.GetSyncedLayer().GetNumber(),
			status.GetTopLayer().GetNumber(),
			status.GetVerifiedLayer().GetNumber(),
		)

		//        switch res.GetStatus() {
		//        case pb.NodeSyncStatus_NOT_SYNCED:
		//            c.syncStart()
		//        }
	}
}

//func (c *Collector) errorPump() error {
//	req := pb.ErrorStreamRequest{}
//
//	log.Info("Start node error pump")
//	defer func() {
//		c.notify <- -streamType_node_Error
//		log.Info("Stop node error pump")
//	}()
//
//	c.notify <- +streamType_node_Error
//
//	stream, err := c.nodeClient.ErrorStream(context.Background(), &req)
//	if err != nil {
//		log.Error("cannot get error stream: %v", err)
//		return err
//	}
//
//	for {
//		res, err := stream.Recv()
//		if err == io.EOF {
//			log.Info("errorPump: EOF")
//			return err
//		}
//		if err != nil {
//			log.Error("cannot receive error: %v", err)
//			return err
//		}
//
//		log.Info("Node error: %v", res.GetError().GetMsg())
//	}
//}
