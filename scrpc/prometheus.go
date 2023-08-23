package scrpc

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var defaultServerMetrics = newServerMetrics()

func init() {
	defaultServerMetrics.mustRegister()
}

type serverMetrics struct {
	totalRequests         *prometheus.CounterVec
	totalRequestsPerRoute *prometheus.CounterVec
	requestLatency        *prometheus.HistogramVec
}

func newServerMetrics() *serverMetrics {
	var s serverMetrics
	s.totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incoming_grpc_requests_total",
			Help: "The number of incoming gRPC requests.",
		},
		[]string{"status"},
	)

	s.totalRequestsPerRoute = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incoming_grpc_requests_per_route_total",
			Help: "The number of incoming gRPC requests per route.",
		},
		[]string{"method", "status"},
	)

	s.requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "incoming_grpc_request_latency_milliseconds",
			Help:    "A histogram of the response latency for gRPC requests in milliseconds.",
			Buckets: prometheus.ExponentialBuckets(25, 2, 7),
		},
		[]string{"method"},
	)

	return &s
}

func (m *serverMetrics) mustRegister() {
	prometheus.MustRegister(m.totalRequests, m.totalRequestsPerRoute, m.requestLatency)
}

func (m *serverMetrics) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		st, _ := status.FromError(err)

		m.requestLatency.With(prometheus.Labels{
			"method": info.FullMethod,
		}).Observe(float64(time.Since(start) / time.Millisecond))

		m.totalRequests.With(prometheus.Labels{
			"status": st.Code().String(),
		}).Inc()

		m.totalRequestsPerRoute.With(prometheus.Labels{
			"method": info.FullMethod,
			"status": st.Code().String(),
		}).Inc()

		return resp, err
	}
}
