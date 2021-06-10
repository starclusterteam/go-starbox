package metrics_test

import (
	"context"
	"fmt"
	"testing"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/starclusterteam/go-starbox/metrics"
	"github.com/stretchr/testify/require"
)

func TestPrometheusServer(t *testing.T) {
	metricsServer := metrics.NewPrometheusServer()
	go metricsServer.Run()
	defer metricsServer.Stop(context.Background())

	testPrometheusServer(t, 13.2, 9990, "/metrics")
}

func TestPrometheusServerCustomPortAndPath(t *testing.T) {
	var (
		port = 10001
		path = "/metrics"
	)

	metricsServer := metrics.NewPrometheusServer(metrics.PrometheusPort(port), metrics.PrometheusPath(path))
	go metricsServer.Run()
	defer metricsServer.Stop(context.Background())

	testPrometheusServer(t, 42, port, path)
}

func testPrometheusServer(t *testing.T, counterValue float64, port int, path string) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "stdlib_prometheus",
		Name:      "test_counter",
		Help:      "Test counter for the stdlib Prometheus webserver",
	})

	prometheus.MustRegister(counter)
	defer prometheus.Unregister(counter)

	counter.Add(counterValue)

	parser := expfmt.TextParser{}
	resp, err := retryablehttp.Get(fmt.Sprintf("http://localhost:%d/%s", port, path))
	require.NoError(t, err)

	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	require.NoError(t, err)

	require.Contains(t, metricFamilies, "stdlib_prometheus_test_counter")
	require.Len(t, metricFamilies["stdlib_prometheus_test_counter"].Metric, 1)
	require.NotNil(t, metricFamilies["stdlib_prometheus_test_counter"].Metric[0].GetCounter())
	require.Equal(t, counterValue, metricFamilies["stdlib_prometheus_test_counter"].Metric[0].GetCounter().GetValue())
}
