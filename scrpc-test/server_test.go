package scrpc_test

import (
	"context"
	"testing"

	"github.com/starclusterteam/go-starbox/scrpc"
	pb "github.com/starclusterteam/go-starbox/scrpc-test/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	serverName = "scrpc-test-grpc.default"
)

type testServer struct {
	pb.UnsafeTestServiceServer
}

func (s *testServer) Test(context.Context, *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func TestServer(t *testing.T) {
	s, err := scrpc.NewServer(func(s *grpc.Server) {
		pb.RegisterTestServiceServer(s, &testServer{})
	})
	require.NoError(t, err)

	// Run grpc server in a separate goroutine.
	var g errgroup.Group
	g.Go(s.Run)

	// Create a connection with the TLS credentials
	conn, err := grpc.Dial("localhost:18444", grpc.WithInsecure())
	require.NoError(t, err)

	// Initialize the client and make the request
	client := pb.NewTestServiceClient(conn)
	_, err = client.Test(context.Background(), &pb.Empty{})
	assert.NoError(t, err)

	// Wait for the grpc server to stop.
	s.GracefulStop()
	err = g.Wait()
	if err != grpc.ErrServerStopped {
		require.NoError(t, err)
	}
}
