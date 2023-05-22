package testserver

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/explorer-backend/test/testseed"
	"google.golang.org/grpc"
	"net"
)

type FakePrivateNode struct {
	seedGen        *testseed.SeedGenerator
	NodePort       int
	InitDone       chan struct{}
	server         *grpc.Server
	smesherService *smesherServiceWrapper
}

// Start register fake services and start stream generated data.
func (f *FakePrivateNode) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", f.NodePort))
	if err != nil {
		return fmt.Errorf("failed to listen fake node: %v", err)
	}

	f.server = grpc.NewServer()
	pb.RegisterSmesherServiceServer(f.server, f.smesherService)
	return f.server.Serve(lis)
}

// Stop stop fake node.
func (f *FakePrivateNode) Stop() {
	f.server.Stop()
}

func (s *smesherServiceWrapper) PostConfig(context.Context, *empty.Empty) (*pb.PostConfigResponse, error) {
	return &pb.PostConfigResponse{
		BitsPerLabel:  s.seed.BitsPerLabel,
		LabelsPerUnit: s.seed.LabelsPerUnit,
		MinNumUnits:   s.seed.MinNumUnits,
		MaxNumUnits:   s.seed.MaxNumUnits,
	}, nil
}
