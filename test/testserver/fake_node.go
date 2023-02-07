package testserver

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/phayes/freeport"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"google.golang.org/genproto/googleapis/rpc/code"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/spacemeshos/explorer-backend/test/testseed"
	"github.com/spacemeshos/explorer-backend/utils"
)

const (
	methodSend = 16
)

type meshServiceWrapper struct {
	startTime time.Time
	seed      *testseed.TestServerSeed
	seedGen   *testseed.SeedGenerator
	pb.UnimplementedMeshServiceServer
}

type debugServiceWrapper struct {
	seedGen *testseed.SeedGenerator
	pb.UnimplementedDebugServiceServer
}

type smesherServiceWrapper struct {
	seedGen *testseed.SeedGenerator
	seed    *testseed.TestServerSeed
	pb.UnimplementedSmesherServiceServer
}

type globalStateServiceWrapper struct {
	seedGen *testseed.SeedGenerator
	pb.UnimplementedGlobalStateServiceServer
}

type nodeServiceWrapper struct {
	seedGen *testseed.SeedGenerator
	pb.UnimplementedNodeServiceServer
}

// FakeNode simulate spacemesh node.
type FakeNode struct {
	seedGen        *testseed.SeedGenerator
	NodePort       int
	InitDone       chan struct{}
	server         *grpc.Server
	nodeService    *nodeServiceWrapper
	meshService    *meshServiceWrapper
	globalState    *globalStateServiceWrapper
	debugService   *debugServiceWrapper
	smesherService *smesherServiceWrapper
}

var stateSynced = make(chan struct{})

// CreateFakeSMNode create a fake spacemesh node.
func CreateFakeSMNode(startTime time.Time, seedGen *testseed.SeedGenerator, seedConf *testseed.TestServerSeed) (*FakeNode, error) {
	appPort, err := freeport.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %v", err)
	}
	return &FakeNode{
		seedGen:        seedGen,
		NodePort:       appPort,
		InitDone:       make(chan struct{}),
		nodeService:    &nodeServiceWrapper{seedGen, pb.UnimplementedNodeServiceServer{}},
		meshService:    &meshServiceWrapper{startTime, seedConf, seedGen, pb.UnimplementedMeshServiceServer{}},
		globalState:    &globalStateServiceWrapper{seedGen, pb.UnimplementedGlobalStateServiceServer{}},
		debugService:   &debugServiceWrapper{seedGen, pb.UnimplementedDebugServiceServer{}},
		smesherService: &smesherServiceWrapper{seedGen, seedConf, pb.UnimplementedSmesherServiceServer{}},
	}, nil
}

// Start register fake services and start stream generated data.
func (f *FakeNode) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", f.NodePort))
	if err != nil {
		return fmt.Errorf("failed to listen fake node: %v", err)
	}

	f.server = grpc.NewServer()
	pb.RegisterDebugServiceServer(f.server, f.debugService)
	pb.RegisterMeshServiceServer(f.server, f.meshService)
	pb.RegisterNodeServiceServer(f.server, f.nodeService)
	pb.RegisterGlobalStateServiceServer(f.server, f.globalState)
	pb.RegisterSmesherServiceServer(f.server, f.smesherService)
	return f.server.Serve(lis)
}

// Stop stop fake node.
func (f *FakeNode) Stop() {
	f.server.Stop()
}

func (m meshServiceWrapper) GenesisTime(context.Context, *pb.GenesisTimeRequest) (*pb.GenesisTimeResponse, error) {
	return &pb.GenesisTimeResponse{Unixtime: &pb.SimpleInt{Value: uint64(m.startTime.Unix())}}, nil
}

func (m meshServiceWrapper) GenesisID(context.Context, *pb.GenesisIDRequest) (*pb.GenesisIDResponse, error) {
	return &pb.GenesisIDResponse{GenesisId: []byte("genesisid")}, nil
}

func (m meshServiceWrapper) EpochNumLayers(context.Context, *pb.EpochNumLayersRequest) (*pb.EpochNumLayersResponse, error) {
	return &pb.EpochNumLayersResponse{Numlayers: &pb.SimpleInt{Value: m.seed.EpochNumLayers}}, nil
}

func (m meshServiceWrapper) LayerDuration(context.Context, *pb.LayerDurationRequest) (*pb.LayerDurationResponse, error) {
	return &pb.LayerDurationResponse{Duration: &pb.SimpleInt{Value: m.seed.LayersDuration}}, nil
}

func (m meshServiceWrapper) MaxTransactionsPerSecond(context.Context, *pb.MaxTransactionsPerSecondRequest) (*pb.MaxTransactionsPerSecondResponse, error) {
	return &pb.MaxTransactionsPerSecondResponse{MaxTxsPerSecond: &pb.SimpleInt{Value: m.seed.MaxTransactionPerSecond}}, nil
}

func (d *debugServiceWrapper) Accounts(context.Context, *empty.Empty) (*pb.AccountsResponse, error) {
	accs := make([]*pb.Account, 0, len(d.seedGen.Accounts))
	for _, acc := range d.seedGen.Accounts {
		accs = append(accs, &pb.Account{
			AccountId: &pb.AccountId{Address: acc.Account.Address},
			StateProjected: &pb.AccountState{
				Balance: &pb.Amount{Value: acc.Account.Balance},
				Counter: acc.Account.Counter,
			},
		})
	}
	return &pb.AccountsResponse{AccountWrapper: accs}, nil
}

func (s *smesherServiceWrapper) PostConfig(context.Context, *empty.Empty) (*pb.PostConfigResponse, error) {
	return &pb.PostConfigResponse{
		BitsPerLabel:  s.seed.BitsPerLabel,
		LabelsPerUnit: s.seed.LabelsPerUnit,
		MinNumUnits:   s.seed.MinNumUnits,
		MaxNumUnits:   s.seed.MaxNumUnits,
	}, nil
}

func (g *globalStateServiceWrapper) GlobalStateStream(request *pb.GlobalStateStreamRequest, stream pb.GlobalStateService_GlobalStateStreamServer) error {
	<-stateSynced
	println("global state stream started")
	for _, epoch := range g.seedGen.Epochs {
		for _, reward := range epoch.Rewards {
			resp := &pb.GlobalStateStreamResponse{Datum: &pb.GlobalStateData{Datum: &pb.GlobalStateData_Reward{
				Reward: &pb.Reward{
					LayerComputed: &pb.LayerNumber{Number: reward.LayerComputed},
					Layer:         &pb.LayerNumber{Number: reward.Layer},
					Total:         &pb.Amount{Value: reward.Total},
					LayerReward:   &pb.Amount{Value: reward.LayerReward},
					Coinbase:      &pb.AccountId{Address: reward.Coinbase},
					Smesher:       &pb.SmesherId{Id: addressToBytes(reward.Smesher)},
				},
			}}}
			if err := stream.Send(resp); err != nil {
				return fmt.Errorf("failed to send global state stream response: %v", err)
			}
		}
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			println("global state stream sending")
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (g *globalStateServiceWrapper) Account(_ context.Context, req *pb.AccountRequest) (*pb.AccountResponse, error) {
	acc, ok := g.seedGen.Accounts[strings.ToLower(req.AccountId.Address)]
	if !ok {
		return nil, status.Error(codes.NotFound, "account not found")
	}
	return &pb.AccountResponse{
		AccountWrapper: &pb.Account{
			AccountId: &pb.AccountId{
				Address: req.AccountId.Address,
			},
			StateCurrent: &pb.AccountState{
				Balance: &pb.Amount{Value: acc.Account.Balance},
				Counter: acc.Account.Counter,
			},
		},
	}, nil
}

func (m *meshServiceWrapper) LayerStream(_ *pb.LayerStreamRequest, stream pb.MeshService_LayerStreamServer) error {
	if err := m.sendEpoch(stream); err != nil {
		return fmt.Errorf("failed to send epoch: %v", err)
	}
	println("sended all layers")
	time.Sleep(1 * time.Second)
	close(stateSynced)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			println("sending layers")
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (m *meshServiceWrapper) sendEpoch(stream pb.MeshService_LayerStreamServer) error {
	for _, epoch := range m.seedGen.Epochs {
		for _, layerContainer := range epoch.Layers {
			atx := make([]*pb.Activation, 0, len(layerContainer.Activations))
			for _, atxGenerated := range layerContainer.Activations {
				atx = append(atx, &pb.Activation{
					Id:        &pb.ActivationId{Id: mustParse(atxGenerated.Id)},
					Layer:     &pb.LayerNumber{Number: atxGenerated.Layer},
					SmesherId: &pb.SmesherId{Id: addressToBytes(atxGenerated.SmesherId)},
					Coinbase:  &pb.AccountId{Address: atxGenerated.Coinbase},
					PrevAtx:   &pb.ActivationId{Id: mustParse(atxGenerated.PrevAtx)},
					NumUnits:  atxGenerated.NumUnits,
				})
			}
			blocksRes := make([]*pb.Block, 0)
			for _, blockContainer := range layerContainer.Blocks {
				tx := make([]*pb.Transaction, 0, len(blockContainer.Transactions))
				for _, txContainer := range blockContainer.Transactions {
					receiver, err := types.StringToAddress(txContainer.Receiver)
					if err != nil {
						panic("invalid receiver address: " + err.Error())
					}
					signer := m.seedGen.Accounts[strings.ToLower(txContainer.Sender)].Signer
					tx = append(tx, &pb.Transaction{
						Id:     mustParse(txContainer.Id),
						Method: methodSend,
						Principal: &pb.AccountId{
							Address: txContainer.Sender,
						},
						GasPrice: txContainer.GasPrice,
						MaxGas:   txContainer.MaxGas,
						Nonce: &pb.Nonce{
							Counter: txContainer.Counter,
						},
						Template: &pb.AccountId{
							Address: wallet.TemplateAddress.String(),
						},
						Raw: sdkWallet.Spend(signer.PrivateKey(), receiver, txContainer.Amount, txContainer.Counter, sdk.WithGasPrice(txContainer.GasPrice)),
					})
				}
				blocksRes = append(blocksRes, &pb.Block{
					Id:           mustParse(blockContainer.Block.Id),
					Transactions: tx,
					SmesherId: &pb.SmesherId{
						Id: addressToBytes(blockContainer.SmesherID),
					},
				})
			}
			pbLayer := &pb.Layer{
				Number:      &pb.LayerNumber{Number: layerContainer.Layer.Number},
				Status:      pb.Layer_LayerStatus(layerContainer.Layer.Status),
				Hash:        mustParse(layerContainer.Layer.Hash),
				Blocks:      blocksRes,
				Activations: atx,
			}
			if err := stream.Send(&pb.LayerStreamResponse{Layer: pbLayer}); err != nil {
				return fmt.Errorf("send to stream: %w", err)
			}
		}
	}
	return nil
}

func (n *nodeServiceWrapper) SyncStart(context.Context, *pb.SyncStartRequest) (*pb.SyncStartResponse, error) {
	return &pb.SyncStartResponse{Status: &rpcstatus.Status{Code: int32(code.Code_OK)}}, nil
}

func (n *nodeServiceWrapper) StatusStream(req *pb.StatusStreamRequest, stream pb.NodeService_StatusStreamServer) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			curLayer, latestLayer, verifiedLayer := n.seedGen.GetLastLayer()
			resp := &pb.StatusStreamResponse{
				Status: &pb.NodeStatus{
					ConnectedPeers: uint64(rand.Intn(10)) + 1,              // number of connected peers
					IsSynced:       true,                                   // whether the node is synced
					SyncedLayer:    &pb.LayerNumber{Number: latestLayer},   // latest layer we saw from the network
					TopLayer:       &pb.LayerNumber{Number: curLayer},      // current layer, based on time
					VerifiedLayer:  &pb.LayerNumber{Number: verifiedLayer}, // latest verified layer
				},
			}

			if err := stream.Send(resp); err != nil {
				return fmt.Errorf("send to stream: %w", err)
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (n *nodeServiceWrapper) Status(context.Context, *pb.StatusRequest) (*pb.StatusResponse, error) {
	return &pb.StatusResponse{Status: &pb.NodeStatus{SyncedLayer: &pb.LayerNumber{Number: 0}}}, nil
}

func mustParse(str string) []byte {
	res, err := utils.StringToBytes(str)
	if err != nil {
		panic("error while parse string to bytes: " + err.Error())
	}
	return res
}

func addressToBytes(addr string) []byte {
	res, err := types.StringToAddress(addr)
	if err != nil {
		panic("error while parse string to address: " + err.Error())
	}
	return res.Bytes()
}
