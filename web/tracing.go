package web

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	zipkin "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/pkg/errors"
	"github.com/starclusterteam/go-starbox/log"
	"github.com/uber/jaeger-client-go"
)

// HTTP server tracing.

// TracingMiddleware returns a middleware that adds tracing information to a request.
func TracingMiddleware(tracer opentracing.Tracer, operationName string) func(http.Handler) http.Handler {
	// If no tracer is given, return a noop middleware.
	if tracer == nil {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := GetLogger(r)

			wireContext, err := tracer.Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.Header),
			)
			if err != nil {
				logger.Debugf("Error encountered while trying to extract span: %v", err)
			}

			span := tracer.StartSpan(operationName, opentracing.ChildOf(wireContext))
			defer span.Finish()

			traceID, err := extractTraceID(span.Context())
			if err != nil {
				logger.Errorf("Failed to extract trace id: %v", err)
			}

			// Annotate logger with trace id.
			SetLogger(r, logger.With("trace_id", traceID))

			// Set trace id in response.
			w.Header().Add("X-Trace-Id", traceID)

			// span tags
			ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
			ext.HTTPUrl.Set(span, r.URL.String())
			ext.HTTPMethod.Set(span, r.Method)
			ext.PeerAddress.Set(span, r.RemoteAddr)

			host, port, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				ip := net.ParseIP(host)

				if ipv4 := ip.To4(); ipv4 != nil {
					ext.PeerHostIPv4.Set(span, binary.BigEndian.Uint32(ipv4))
				} else {
					ext.PeerHostIPv6.Set(span, ip.String())
				}

				uintPort, err := strconv.ParseUint(port, 10, 16)
				if err == nil {
					ext.PeerPort.Set(span, uint16(uintPort))
				}
			}

			// store span in context
			ctx := opentracing.ContextWithSpan(r.Context(), span)

			// update request context to include the new span
			r = r.WithContext(ctx)

			// next middleware or actual request handler
			next.ServeHTTP(w, r)
		})
	}
}

func extractTraceID(spanCtx opentracing.SpanContext) (string, error) {
	var traceID string

	switch spanCtxImpl := spanCtx.(type) {
	case zipkin.SpanContext:
		traceID = convertTraceID(spanCtxImpl.TraceID.Low, spanCtxImpl.TraceID.High)
	case mocktracer.MockSpanContext:
		traceID = strconv.Itoa(spanCtxImpl.TraceID)
	case jaeger.SpanContext:
		traceID = spanCtxImpl.TraceID().String()
	default:
		traceID = "" // maybe opentracing.noopSpanContext ??
	}

	return traceID, nil
}

func convertTraceID(traceID uint64, traceIDHigh uint64) string {
	buf := bytes.NewBuffer([]byte{})
	if traceIDHigh != 0 {
		binary.Write(buf, binary.BigEndian, &traceIDHigh)
	}
	binary.Write(buf, binary.BigEndian, &traceID)
	res := ""
	for _, b := range buf.Bytes() {
		res = res + fmt.Sprintf("%02x", b)
	}
	return res
}

// HTTP client tracing.

type transport struct {
	operationName string
	transport     http.RoundTripper
	tracer        opentracing.Tracer
}

// TracedTransport takes a http.RoundTripper and returns a transport wrapper
// that implements RoundTrip with added tracing.
func TracedTransport(tracer opentracing.Tracer, t http.RoundTripper, operationName string) http.RoundTripper {
	return &transport{
		operationName: operationName,
		transport:     t,
		tracer:        tracer,
	}
}

// RoundTrip adds tracing information to request before calling the underlying
// RoundTripper.
func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	span := t.tracer.StartSpan(t.operationName)
	defer span.Finish()

	tracedReq, err := TraceableRequestFunc(t.tracer)(span, r)
	if err != nil {
		log.Errorf("failed to create traceable request")
		tracedReq = r
	}

	// fallback to DefaultTransport if no transport was given
	if t.transport == nil {
		return http.DefaultTransport.RoundTrip(tracedReq)
	}

	return t.transport.RoundTrip(tracedReq)
}

// RequestFunc defines a function that receives a span and inserts it into the
// request context.
type RequestFunc func(opentracing.Span, *http.Request) (*http.Request, error)

// TraceableRequestFunc returns a RequestFunc.
func TraceableRequestFunc(tracer opentracing.Tracer) RequestFunc {
	return func(span opentracing.Span, r *http.Request) (*http.Request, error) {
		if span == nil {
			return r, nil
		}

		// add standard OpenTracing tags
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())
		ext.PeerHostname.Set(span, r.URL.Host)

		// inject the span context into the outgoing HTTP request
		err := tracer.Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)
		return r, errors.Wrap(err, "failed to inject span in request")
	}
}
