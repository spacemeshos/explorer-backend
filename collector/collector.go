package collector

import (
	"context"
	"errors"
	"github.com/spacemeshos/explorer-backend/collector/sql"
	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/go-spacemesh/common/types"
	sql2 "github.com/spacemeshos/go-spacemesh/sql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/keepalive"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"google.golang.org/grpc"

	"github.com/spacemeshos/go-spacemesh/log"
)

const (
	streamType_node_SyncStatus int = 1
	//streamType_mesh_Layer       int = 2
	streamType_transactions     int = 2
	streamType_mesh_Malfeasance int = 3

	streamType_count int = 3
)

type Listener interface {
	OnNetworkInfo(genesisId string, genesisTime uint64, epochNumLayers uint32, maxTransactionsPerSecond uint64, layerDuration uint64, postUnitSize uint64)
	OnNodeStatus(connectedPeers uint64, isSynced bool, syncedLayer uint32, topLayer uint32, verifiedLayer uint32)
	OnLayer(layer *pb.Layer)
	OnAccounts(accounts []*types.Account)
	OnReward(reward *pb.Reward)
	OnMalfeasanceProof(proof *pb.MalfeasanceProof)
	OnTransactionResult(res *pb.TransactionResult, state *pb.TransactionState)
	GetLastLayer(parent context.Context) uint32
	LayersInQueue() int
	IsLayerInQueue(layer *pb.Layer) bool
	GetEpochNumLayers() uint32
	GetTransactions(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]model.Transaction, error)
	UpdateTransactionState(parent context.Context, id string, state int32) error
	UpdateEpochStats(layer uint32)
	OnActivation(atx *types.ActivationTx)
	GetLastActivationReceived() int64
	RecalculateEpochStats()
	OnActivations(atxs []*model.Activation)
}

type Collector struct {
	apiPublicUrl              string
	apiPrivateUrl             string
	syncMissingLayersFlag     bool
	recalculateEpochStatsFlag bool
	syncFromLayerFlag         uint32
	atxSyncFlag               bool

	listener Listener
	db       *sql2.Database
	dbClient sql.DatabaseClient

	nodeClient         pb.NodeServiceClient
	meshClient         pb.MeshServiceClient
	globalClient       pb.GlobalStateServiceClient
	transactionsClient pb.TransactionServiceClient
	debugClient        pb.DebugServiceClient
	smesherClient      pb.SmesherServiceClient

	streams       [streamType_count]bool
	activeStreams int
	connecting    bool
	online        bool
	closing       bool

	// Stream status changed.
	notify chan int
}

func NewCollector(nodePublicAddress string, nodePrivateAddress string, syncMissingLayersFlag bool,
	syncFromLayerFlag int, recalculateEpochStatsFlag bool,
	listener Listener, db *sql2.Database, dbClient sql.DatabaseClient, atxSyncFlag bool) *Collector {
	return &Collector{
		apiPublicUrl:              nodePublicAddress,
		apiPrivateUrl:             nodePrivateAddress,
		syncMissingLayersFlag:     syncMissingLayersFlag,
		recalculateEpochStatsFlag: recalculateEpochStatsFlag,
		syncFromLayerFlag:         uint32(syncFromLayerFlag),
		listener:                  listener,
		notify:                    make(chan int),
		db:                        db,
		dbClient:                  dbClient,
		atxSyncFlag:               atxSyncFlag,
	}
}

func (c *Collector) Run() error {
	log.Info("dial node %v and %v", c.apiPublicUrl, c.apiPrivateUrl)
	c.connecting = true

	//TODO: move to env
	keepaliveOpts := keepalive.ClientParameters{
		Time:                4 * time.Minute,
		Timeout:             2 * time.Minute,
		PermitWithoutStream: true,
	}

	publicConn, err := grpc.Dial(c.apiPublicUrl, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepaliveOpts),
		grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024)))
	if err != nil {
		return errors.Join(errors.New("cannot dial node"), err)
	}
	defer publicConn.Close()

	privateConn, err := grpc.Dial(c.apiPrivateUrl, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepaliveOpts),
		grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024)))
	if err != nil {
		return errors.Join(errors.New("cannot dial node"), err)
	}
	defer privateConn.Close()

	c.nodeClient = pb.NewNodeServiceClient(publicConn)
	c.meshClient = pb.NewMeshServiceClient(publicConn)
	c.globalClient = pb.NewGlobalStateServiceClient(publicConn)
	c.transactionsClient = pb.NewTransactionServiceClient(publicConn)
	c.debugClient = pb.NewDebugServiceClient(publicConn)
	c.smesherClient = pb.NewSmesherServiceClient(privateConn)

	err = c.getNetworkInfo()
	if err != nil {
		return errors.Join(errors.New("cannot get network info"), err)
	}

	if c.atxSyncFlag {
		err = c.syncActivations()
		if err != nil {
			return errors.Join(errors.New("cannot sync activations"), err)
		}
	}

	if c.syncMissingLayersFlag {
		err = c.syncMissingLayers()
		if err != nil {
			return errors.Join(errors.New("cannot sync missing layers"), err)
		}
	}

	if c.recalculateEpochStatsFlag {
		c.listener.RecalculateEpochStats()
	}

	g := new(errgroup.Group)
	g.Go(func() error {
		err := c.syncStatusPump()
		if err != nil {
			return errors.Join(errors.New("cannot start sync status pump"), err)
		}
		return nil
	})

	g.Go(func() error {
		err := c.transactionsPump()
		if err != nil {
			return errors.Join(errors.New("cannot start transactions pump"), err)
		}
		return nil
	})

	g.Go(func() error {
		err := c.malfeasancePump()
		if err != nil {
			return errors.Join(errors.New("cannot start sync malfeasance pump"), err)
		}
		return nil
	})

	g.Go(func() error {
		for c.connecting || c.closing || c.online {
			state := <-c.notify
			log.Info("stream notify %v", state)
			switch {
			case state > 0:
				c.streams[state-1] = true
				c.activeStreams++
				log.Info("stream connected %v", state)
			case state < 0:
				c.streams[(-state)-1] = false
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
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
