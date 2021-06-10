package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/starclusterteam/go-starbox/config"
	"github.com/starclusterteam/go-starbox/constants/envvar"
)

var defaultServerMetrics = NewServerMetrics()

func init() {
	if config.Bool(envvar.PrometheusEnabled, false) {
		defaultServerMetrics.mustRegister()
	}
}

type serverMetrics struct {
	totalRequests         *prometheus.CounterVec
	totalRequestsPerRoute *prometheus.CounterVec
	requestLatency        *prometheus.HistogramVec
}

func NewServerMetrics() *serverMetrics {
	var s serverMetrics
	s.totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incoming_http_requests_total",
			Help: "The number of incoming HTTP requests.",
		},
		[]string{"status", "statusClass"},
	)

	s.totalRequestsPerRoute = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "incoming_http_requests_per_route_total",
			Help: "The number of incoming HTTP requests per route.",
		},
		[]string{"method", "status", "statusClass", "url"},
	)

	s.requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "incoming_http_request_latency_milliseconds",
			Help:    "A histogram of the response latency for HTTP requests in milliseconds.",
			Buckets: []float64{50, 100, 200, 400, 800, 1600, 3200},
		},
		[]string{"method", "url"},
	)

	return &s
}

func (m *serverMetrics) mustRegister() {
	prometheus.MustRegister(m.totalRequests, m.totalRequestsPerRoute, m.requestLatency)
}

// Middleware reports request duration, method and status code to prometheus.
func (m *serverMetrics) Middleware(pattern string) func(next http.Handler) http.Handler {
	totalRequestsPerRoute := m.totalRequestsPerRoute.MustCurryWith(prometheus.Labels{"url": pattern})
	requestLatency := m.requestLatency.MustCurryWith(prometheus.Labels{"url": pattern})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			requestLatency.With(prometheus.Labels{
				"method": r.Method,
			}).Observe(float64(metrics.Duration / time.Millisecond))

			m.totalRequests.With(prometheus.Labels{
				"status":      statusCodeToString(metrics.Code),
				"statusClass": statusCodeToClass(metrics.Code),
			}).Inc()

			totalRequestsPerRoute.With(prometheus.Labels{
				"method":      r.Method,
				"status":      statusCodeToString(metrics.Code),
				"statusClass": statusCodeToClass(metrics.Code),
			}).Inc()
		})
	}
}

func statusCodeToString(code int) string {
	// status defaults to "200" in case code is 0, conforming to the http standard library
	if code == 0 {
		return "200"
	}

	return strconv.Itoa(code)
}

func statusCodeToClass(code int) string {
	// status defaults to "200" in case code is 0, conforming to the http standard library
	if code == 0 {
		return "2xx"
	}

	status := strconv.Itoa(code)
	if len(status) == 0 {
		return ""
	}
	return status[:1] + "xx"
}
