package scrpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/starclusterteam/go-starbox/config"
	"github.com/starclusterteam/go-starbox/log"
	"github.com/starclusterteam/go-starbox/tracing"

	// This metrics import is used to initialize Prometheus HTTP endpoint server.
	_ "github.com/starclusterteam/go-starbox/metrics/auto"
)

const (
	defaultPort    = 18444
	defaultTLSPort = 18555
)

const (
	portEnv          = "GRPC_PORT"
	tlsPortEnv       = "GRPC_SSL_PORT"
	tlsClientCAsEnv  = "GRPC_CLIENT_CA"
	tlsServerCertEnv = "GRPC_SERVER_CERT"
	tlsServerKeyEnv  = "GRPC_SERVER_KEY"
)

// Server encapsulates gRPC server
type Server struct {
	addr   string
	server *grpc.Server
}

// NewServer creates a new server for gRPC
func NewServer(cb func(*grpc.Server), opts ...ServerOption) (*Server, error) {
	options := options{
		addr:    fmt.Sprintf(":%d", config.Int(portEnv, defaultPort)),
		tlsAddr: fmt.Sprintf(":%d", config.Int(tlsPortEnv, defaultTLSPort)),
		tracer:  tracing.Tracer,
	}

	for _, o := range opts {
		o(&options)
	}

	creds, err := resolveTLSConfig(options.tlsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve TLS config")
	}

	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.ChainUnaryInterceptor(
			tracingInterceptor(options.tracer, options.traceHealthCheck),
			LoggerInterceptor,
			defaultServerMetrics.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(),
		),
		grpc.Creds(creds),
	)

	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// Run callback to add services
	cb(server)

	var addr string
	if options.tlsConfig != nil {
		addr = options.tlsAddr
	} else {
		addr = options.addr
	}

	return &Server{
		server: server,
		addr:   addr,
	}, nil
}

// Run starts gRPC server.
func (s *Server) Run() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Infof("Running gRPC server on %s", l.Addr())
	return s.server.Serve(l)
}

// GracefulStop stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}

func tracingInterceptor(tracer opentracing.Tracer, traceHealthCheck bool) grpc.UnaryServerInterceptor {
	if traceHealthCheck {
		return otgrpc.OpenTracingServerInterceptor(tracer)
	}

	return otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.IncludingSpans(excludeHealthCheckFromTrace()))
}

func resolveTLSConfig(tlsConfig *serverTLSConfig) (credentials.TransportCredentials, error) {
	if tlsConfig == nil {
		return nil, nil
	}

	var (
		keyFile       = tlsConfig.keyFile
		certFile      = tlsConfig.certFile
		clientCaFiles = tlsConfig.clientCaFiles
	)

	// Load the certificates from disk.
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load server key pair")
	}

	// Create a certificate pool from the certificate authority.
	certPool := x509.NewCertPool()
	for _, clientCaFile := range clientCaFiles {
		c, err := ioutil.ReadFile(clientCaFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read ca cert")
		}

		// Append the client certificates from the CA.
		if ok := certPool.AppendCertsFromPEM(c); !ok {
			return nil, errors.Wrap(err, "failed to append ca certs")
		}
	}

	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	return creds, nil
}

type serverTLSConfig struct {
	certFile      string
	keyFile       string
	clientCaFiles []string
}

type options struct {
	traceHealthCheck bool
	addr             string
	tlsAddr          string
	tracer           opentracing.Tracer
	tlsConfig        *serverTLSConfig
}

// ServerOption define a functional options used when creating a grpc server.
type ServerOption func(*options)

// TraceHealthCheck is used to set if the grpc health check is traced or not. It defaults to false.
func TraceHealthCheck(b bool) ServerOption {
	return func(o *options) {
		o.traceHealthCheck = b
	}
}

// WithPort sets the listening address for the server as ":<port>".
func WithPort(port int) ServerOption {
	return func(o *options) {
		o.addr = fmt.Sprintf(":%d", port)
		o.tlsAddr = fmt.Sprintf(":%d", port)
	}
}

// WithAddr sets the listening address for the server.
func WithAddr(addr string) ServerOption {
	return func(o *options) {
		o.addr = addr
		o.tlsAddr = addr
	}
}

// WithServerTracer can be used to override the default tracer.
func WithServerTracer(tracer opentracing.Tracer) ServerOption {
	return func(o *options) {
		o.tracer = tracer
	}
}

// WithServerTLSFromParams enables TLS on the server using the given paths for the certificate, key and client CAs.
func WithServerTLSFromParams(certFile, keyFile string, clientCaFiles []string) ServerOption {
	return func(o *options) {
		o.tlsConfig = &serverTLSConfig{
			certFile:      certFile,
			keyFile:       keyFile,
			clientCaFiles: clientCaFiles,
		}
	}
}

// WithServerTLS is similar with WithServerTLSFromParams, only it reads the server certificate, server key and client CAs paths
// from the GRPC_SERVER_CERT, GRPC_SERVER_KEY and GRPC_CLIENT_CA environment variables. If one of the environment variables is not
// set, it will exit the program.
func WithServerTLS() ServerOption {
	tlsServerCert := config.String(tlsServerCertEnv, "")
	if tlsServerCert == "" {
		log.Fatalf("%s must be set", tlsServerCertEnv)
	}

	tlsServerKey := config.String(tlsServerKeyEnv, "")
	if tlsServerKey == "" {
		log.Fatalf("%s must be set", tlsServerKeyEnv)
	}

	tlsClientCAs := strings.Split(config.String(tlsClientCAsEnv, ""), ",")
	if len(tlsClientCAs) == 0 {
		log.Fatalf("%s must be set", tlsClientCAsEnv)
	}

	return WithServerTLSFromParams(tlsServerCert, tlsServerKey, tlsClientCAs)
}

func excludeHealthCheckFromTrace() otgrpc.SpanInclusionFunc {
	return func(
		parentSpanCtx opentracing.SpanContext,
		method string,
		req, resp interface{}) bool {

		if method == "/grpc.health.v1.Health/Check" {
			return false
		}
		return true
	}
}
