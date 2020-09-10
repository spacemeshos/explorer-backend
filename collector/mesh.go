package collector

import (
    "context"
    "io"
    "time"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/go-spacemesh/log"
//    "github.com/spacemeshos/explorer-backend/utils"
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

    netId, err := c.meshClient.NetID(ctx, &pb.NetIDRequest{})
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

    c.listener.OnNetworkInfo(
        netId.GetNetid().GetValue(),
        genesisTime.GetUnixtime().GetValue(),
        epochNumLayers.GetNumlayers().GetValue(),
        maxTransactionsPerSecond.GetMaxtxpersecond().GetValue(),
        layerDuration.GetDuration().GetValue(),
    )

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

    return nil
}
