package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/starclusterteam/go-starbox/config"
	"github.com/starclusterteam/go-starbox/log"
	"github.com/uber/jaeger-client-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

// Tracer is a global opentracing tracer.
// If the ENABLE_TRACING environment variable is set to false, it will default to a noop tracer.
var Tracer opentracing.Tracer = &opentracing.NoopTracer{}

func init() {
	if !config.Bool("TRACING_ENABLED", true) {
		return
	}

	tracer, err := NewTracer()
	if err != nil {
		log.Warningf("Failed to initialize tracer: %v", err)
	} else {
		Tracer = tracer
	}

	opentracing.SetGlobalTracer(Tracer)
}

// NewTracer returns an jaeger implementation of an opentracing tracer.
func NewTracer() (opentracing.Tracer, error) {
	cfg, err := jaegerconfig.FromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build jaeger config from env")
	}

	if cfg.Sampler.Type == "" {
		cfg.Sampler.Type = jaeger.SamplerTypeProbabilistic
		if cfg.Sampler.Param == 0 {
			cfg.Sampler.Param = 0.01
		}
	}

	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	tracer, _, err := cfg.NewTracer(
		jaegerconfig.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegerconfig.Injector(opentracing.TextMap, zipkinPropagator),
		jaegerconfig.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegerconfig.Extractor(opentracing.TextMap, zipkinPropagator),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize jaeger tracer")
	}

	return tracer, nil
}
