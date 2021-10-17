package collector

import (
    "time"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "google.golang.org/grpc"

    "github.com/spacemeshos/go-spacemesh/log"
)

const (
    streamType_node_SyncStatus			int = 1
    streamType_mesh_Layer			int = 2
    streamType_globalState			int = 3
    streamType_node_Error			int = 4

    streamType_count				int = 3
)

type Listener interface {
    OnNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64, postUnitSize uint64)
    OnNodeStatus(connectedPeers uint64, isSynced bool, syncedLayer uint32, topLayer uint32, verifiedLayer uint32)
    OnLayer(layer *pb.Layer)
    OnAccount(account *pb.Account)
    OnReward(reward *pb.Reward)
    OnTransactionReceipt(receipt *pb.TransactionReceipt)
}

type Collector struct {
    apiUrl	string
    listener	Listener

    nodeClient		pb.NodeServiceClient
    meshClient		pb.MeshServiceClient
    globalClient	pb.GlobalStateServiceClient
    debugClient		pb.DebugServiceClient
    smesherClient	pb.SmesherServiceClient

    streams [streamType_count]bool
    activeStreams int
    connecting bool
    online bool
    closing bool

    // Stream status changed.
    notify chan int
}

func NewCollector(nodeAddress string, listener Listener) *Collector {
    return &Collector{
        apiUrl:  nodeAddress,
        listener: listener,
        notify:  make(chan int),
    }
}

func (c *Collector) Run() {
    for {
        log.Info("dial node %v", c.apiUrl)
        c.connecting = true

        conn, err := grpc.Dial(c.apiUrl, grpc.WithInsecure())
        if err != nil {
            log.Error("cannot dial node: %v", err)
            time.Sleep(1 * time.Second)
            continue
        }

        c.nodeClient = pb.NewNodeServiceClient(conn)
        c.meshClient = pb.NewMeshServiceClient(conn)
        c.globalClient = pb.NewGlobalStateServiceClient(conn)
        c.debugClient = pb.NewDebugServiceClient(conn)
        c.smesherClient = pb.NewSmesherServiceClient(conn)

        err = c.getNetworkInfo()
        if err != nil {
            log.Error("cannot get network info: %v", err)
            time.Sleep(1 * time.Second)
            continue
        }

        go c.syncStatusPump()
//        go c.errorPump()
        go c.layersPump()
        go c.globalStatePump()

        for ; c.connecting || c.closing || c.online; {
            state := <-c.notify
            log.Info("stream notify %v", state)
            switch {
            case state > 0:
                c.streams[state - 1] = true
                c.activeStreams++
                log.Info("stream connected %v", state)
            case state < 0:
                c.streams[(-state) - 1] = false
                c.activeStreams--
                if c.activeStreams == 0 {
                    c.closing = false
                }
                log.Info("stream disconnected %v", state)
            }
            if c.activeStreams == streamType_count {
                c.connecting = false
                c.online = true
                log.Info("all streams synchronized!")
            }
            if c.online && c.activeStreams < streamType_count {
                log.Info("streams desynchronized!!!")
                c.online = false
                c.closing = true
                conn.Close()
            }
        }

        log.Info("Wait 1 second...")
        time.Sleep(1 * time.Second)
    }
}
