package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/starclusterteam/go-starbox/log"
)

const (
	defaultPrometheusPort = 9990
	defaultPrometheusPath = "/metrics"
)

type prometheusServer struct {
	*http.Server
	port int
	path string
}

// NewPrometheusServer returns a metrics.Server that, when run, exposes an endpoint with Prometheus-specific metrics.
func NewPrometheusServer(opts ...PrometheusOption) Server {
	s := prometheusServer{
		port: defaultPrometheusPort,
		path: defaultPrometheusPath,
	}

	for _, o := range opts {
		o(&s)
	}

	mux := http.NewServeMux()
	mux.Handle(s.path, promhttp.Handler())

	s.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return &s
}

// Run starts an http server that exposes the Prometheus metrics.
func (s *prometheusServer) Run() error {
	log.Infof("Running Prometheus HTTP endpoint on %s", s.Server.Addr)
	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "failed to start prometheus http server")
	}
	return nil
}

// Stops tries to gracefully shutdown the Prometheus http server.
func (s *prometheusServer) Stop(ctx context.Context) error {
	if err := s.Server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shutdown prometheus http server")
	}
	return nil
}

// PrometheusOption defines a functional option type for the Prometheus webserver config.
type PrometheusOption func(*prometheusServer)

// PrometheusPort is a functional option for setting the Prometheus webserver listening port. Default to 9990.
func PrometheusPort(port int) PrometheusOption {
	return func(s *prometheusServer) {
		s.port = port
	}
}

// PrometheusPath is a functional option for setting the Prometheus webserver metrics path. Defaults to "/metrics".
func PrometheusPath(path string) PrometheusOption {
	return func(s *prometheusServer) {
		s.path = path
	}
}
