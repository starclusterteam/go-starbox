package autometrics

import (
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/starclusterteam/go-starbox/config"
	"github.com/starclusterteam/go-starbox/const/envvar"
	"github.com/starclusterteam/go-starbox/log"
	"github.com/starclusterteam/go-starbox/metrics"
)

const (
	prometheusPortEnv = "PROMETHEUS_TARGET_PORT"
	prometheusPathEnv = "PROMETHEUS_TARGET_PATH"
)

func init() {
	if !config.Bool(envvar.PrometheusEnabled, false) {
		return
	}

	opts, err := resolveOptions()
	if err != nil {
		log.Errorf("Failed to resolve prometheus config: %v", err)
		return
	}

	s := metrics.NewPrometheusServer(opts...)
	go func() {
		if err := s.Run(); err != nil && errors.Cause(err) != http.ErrServerClosed {
			log.Warningf("Failed to run prometheus web server: %v", err)
		}
	}()
}

// resolveOptions looks at the PROMETHEUS_TARGET_PORT and PROMETHEUS_TARGET_PATH environment variables
// to determine the prometheus endpoint configuration. It returns a list of PrometheusOptions
// accepted by the NewPrometheusServer function.
func resolveOptions() ([]metrics.PrometheusOption, error) {
	var opts []metrics.PrometheusOption

	if prometheusPortStr, ok := os.LookupEnv(prometheusPortEnv); ok {
		prometheusPort, err := strconv.Atoi(prometheusPortStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse environment variable(%s)", prometheusPortEnv)
		}

		opts = append(opts, metrics.PrometheusPort(prometheusPort))
	}

	if prometheusPath, ok := os.LookupEnv(prometheusPathEnv); ok {
		opts = append(opts, metrics.PrometheusPath(prometheusPath))
	}

	return opts, nil
}
