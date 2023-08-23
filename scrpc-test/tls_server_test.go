package scrpc_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"

	"github.com/starclusterteam/go-starbox/scrpc"
	pb "github.com/starclusterteam/go-starbox/scrpc-test/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	serverCert = "testdata/certs/server.pem"
	serverKey  = "testdata/certs/server-key.pem"
	serverCa   = "testdata/certs/server-ca.pem"

	client1Ca   = "testdata/certs/client1-ca.pem"
	client1Cert = "testdata/certs/client1.pem"
	client1Key  = "testdata/certs/client1-key.pem"

	client2Ca   = "testdata/certs/client2-ca.pem"
	client2Cert = "testdata/certs/client2.pem"
	client2Key  = "testdata/certs/client2-key.pem"
)

func TestServerWithTLS(t *testing.T) {
	s, err := scrpc.NewServer(func(s *grpc.Server) {
		pb.RegisterTestServiceServer(s, &testServer{})
	}, scrpc.WithServerTLSFromParams(serverCert, serverKey, []string{client1Ca, client2Ca}))
	require.NoError(t, err)

	// Run grpc server in a separate goroutine.
	var g errgroup.Group
	g.Go(s.Run)

	client1 := dialTLSClient(t, "localhost:18555", client1Cert, client1Key, serverCa)
	client2 := dialTLSClient(t, "localhost:18555", client2Cert, client2Key, serverCa)

	_, err = client1.Test(context.Background(), &pb.Empty{})
	assert.NoError(t, err)

	_, err = client2.Test(context.Background(), &pb.Empty{})
	assert.NoError(t, err)

	// Wait for the grpc server to stop.
	s.GracefulStop()
	err = g.Wait()
	if err != grpc.ErrServerStopped {
		require.NoError(t, err)
	}
}

func TestClientWithTLS(t *testing.T) {
	s, err := scrpc.NewServer(func(s *grpc.Server) {
		pb.RegisterTestServiceServer(s, &testServer{})
	}, scrpc.WithServerTLSFromParams(serverCert, serverKey, []string{client1Ca, client2Ca}))
	require.NoError(t, err)

	// Run grpc server in a separate goroutine.
	var g errgroup.Group
	g.Go(s.Run)

	client1 := scrpcDialTLSClient(t, "localhost:18555", client1Cert, client1Key, serverCa)
	client2 := scrpcDialTLSClient(t, "localhost:18555", client2Cert, client2Key, serverCa)

	_, err = client1.Test(context.Background(), &pb.Empty{})
	assert.NoError(t, err)

	_, err = client2.Test(context.Background(), &pb.Empty{})
	assert.NoError(t, err)

	// Wait for the grpc server to stop.
	s.GracefulStop()
	err = g.Wait()
	if err != grpc.ErrServerStopped {
		require.NoError(t, err)
	}
}

func dialTLSClient(t *testing.T, addr, clientCert, clientKey, serverCa string) pb.TestServiceClient {
	// Load the client certificates from disk.
	certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
	require.NoError(t, err)

	// Create a certificate pool from the certificate authority.
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(serverCa)
	require.NoError(t, err)

	// Append the certificates from the CA.
	ok := certPool.AppendCertsFromPEM(ca)
	require.True(t, ok)

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	// Create a connection with the TLS credentials
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	require.NoError(t, err)

	return pb.NewTestServiceClient(conn)
}

func scrpcDialTLSClient(t *testing.T, addr, clientCert, clientKey, serverCa string) pb.TestServiceClient {
	conn, err := scrpc.Dial(addr, scrpc.WithClientTLSFromParams(serverName, clientCert, clientKey, serverCa))
	require.NoError(t, err)
	return pb.NewTestServiceClient(conn)
}
