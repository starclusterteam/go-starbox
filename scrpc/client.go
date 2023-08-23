package scrpc

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/starclusterteam/go-starbox/config"
	"github.com/starclusterteam/go-starbox/tracing"
)

const (
	tlsClientCertEnv = "GRPC_CLIENT_CERT"
	tlsClientKeyEnv  = "GRPC_CLIENT_KEY"
)

// Dial wraps grpc.Dial with the possibility of overriding tracing or adding retry interceptor.
// It is possible to add grpc.DialOptions with WithDialOptions().
func Dial(target string, opts ...DialOption) (*grpc.ClientConn, error) {
	options := dialOptions{
		tracer: tracing.Tracer,
	}

	for _, o := range opts {
		o(&options)
	}

	unaryInterceptor := grpc_middleware.ChainUnaryClient(
		otgrpc.OpenTracingClientInterceptor(options.tracer, options.tracingOpts...),
		grpc_retry.UnaryClientInterceptor(options.retryOpts...),
	)

	tc, err := resolveTransportCredentials(options.tlsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve transport credentials")
	}

	grpcOptions := []grpc.DialOption{grpc.WithUnaryInterceptor(unaryInterceptor), grpc.WithTransportCredentials(tc)}
	grpcOptions = append(grpcOptions, options.opts...)

	return grpc.Dial(target, grpcOptions...)
}

func resolveTransportCredentials(c *clientTLSConfig) (credentials.TransportCredentials, error) {
	if c == nil {
		return nil, nil
	}

	// Load the client certificates from disk.
	certificate, err := tls.LoadX509KeyPair(c.certPath, c.keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load server key pair")
	}

	// Create a certificate pool from the certificate authority.
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(c.serverCAPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read ca cert")
	}

	// Append the certificates from the CA.
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.Wrap(err, "failed to append ca cert")
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   c.serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	return creds, nil
}

// DialOption represents a functional option for Dial.
type DialOption func(*dialOptions)

type dialOptions struct {
	opts []grpc.DialOption

	tracer      opentracing.Tracer
	tracingOpts []otgrpc.Option

	retryOpts []grpc_retry.CallOption

	tlsConfig *clientTLSConfig
}

type clientTLSConfig struct {
	serverName   string
	certPath     string
	keyPath      string
	serverCAPath string
}

// WithTracer is a grpc dial option that adds a tracing middleware to the requests.
func WithTracer(tracer opentracing.Tracer, opts ...otgrpc.Option) DialOption {
	return func(o *dialOptions) {
		o.tracer = tracer
		o.tracingOpts = opts
	}
}

// WithRetry is a grpc dial option that adds retry logic to the requests.
func WithRetry(opts ...grpc_retry.CallOption) DialOption {
	return func(o *dialOptions) {
		o.retryOpts = opts
	}
}

// WithDialOptions sets the grpc.DialOptions for the underlying grpc.Dial.
func WithDialOptions(opts ...grpc.DialOption) DialOption {
	return func(o *dialOptions) {
		o.opts = opts
	}
}

// WithClientTLSFromParams sets the transport credentials on dial.
func WithClientTLSFromParams(serverName, certPath, keyPath, serverCAPath string) DialOption {
	return func(o *dialOptions) {
		o.tlsConfig = &clientTLSConfig{
			serverName:   serverName,
			certPath:     certPath,
			keyPath:      keyPath,
			serverCAPath: serverCAPath,
		}
	}
}

// WithClientTLS is similar to WithClientTLSFromParams, but reads client certificate and key paths
// from the GRPC_CLIENT_CERT and GRPC_CLIENT_KEY environment variables.
func WithClientTLS(serverName, serverCAPath string) DialOption {
	clientCertPath := config.String(tlsClientCertEnv, "")
	clientKeyPath := config.String(tlsClientKeyEnv, "")

	return WithClientTLSFromParams(serverName, clientCertPath, clientKeyPath, serverCAPath)
}
