package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/starclusterteam/go-starbox/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/jaeger-client-go"
)

func TestTracingMiddleware(t *testing.T) {
	var (
		method   = "PUT"
		url      = "/dummy"
		tracer   = mocktracer.New()
		finalReq *http.Request
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalReq = r
	})

	r := httptest.NewRequest(method, url, nil)
	w := httptest.NewRecorder()

	parentSpan := tracer.StartSpan("test.client")
	defer parentSpan.Finish()
	parentSpan.SetBaggageItem("test.baggage", "b")

	tracer.Inject(parentSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

	web.TracingMiddleware(tracer, "tracing-test")(next).ServeHTTP(w, r)

	require.Len(t, tracer.FinishedSpans(), 1)
	span := tracer.FinishedSpans()[0]

	assert.Equal(t, ext.SpanKindEnum("server"), span.Tags()[string(ext.SpanKind)])
	assert.Equal(t, url, span.Tags()[string(ext.HTTPUrl)])
	assert.Equal(t, method, span.Tags()[string(ext.HTTPMethod)])

	// Check if the trace id header was set.
	assert.NotEmpty(t, w.Header().Get("X-Trace-Id"))

	// Check if the baggage was propagated from the parent.
	assert.Equal(t, "b", span.BaggageItem("test.baggage"))

	// Check if the parent span is correctly set.
	assert.Equal(t, parentSpan.(*mocktracer.MockSpan).SpanContext.SpanID, span.ParentID)

	// Check if the span propagates through request context.
	assert.Equal(t, span, opentracing.SpanFromContext(finalReq.Context()))
}

type jaegerMockTracer struct {
	*mocktracer.MockTracer
}

type jaegerMockSpan struct {
	*mocktracer.MockSpan
}

func (t *jaegerMockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	span := t.MockTracer.StartSpan(operationName, opts...)
	return &jaegerMockSpan{span.(*mocktracer.MockSpan)}
}

func (s *jaegerMockSpan) Context() opentracing.SpanContext {
	return jaeger.NewSpanContext(
		jaeger.TraceID{High: 0, Low: uint64(s.MockSpan.SpanContext.TraceID)},
		jaeger.SpanID(s.MockSpan.SpanContext.SpanID),
		jaeger.SpanID(s.ParentID),
		s.MockSpan.SpanContext.Sampled,
		s.MockSpan.SpanContext.Baggage,
	)
}

func TestTracingMiddlewareJaegerSpan(t *testing.T) {
	var (
		method = "PUT"
		url    = "/dummy"
		tracer = &jaegerMockTracer{mocktracer.New()}
	)

	r := httptest.NewRequest(method, url, nil)
	w := httptest.NewRecorder()

	parentSpan := tracer.MockTracer.StartSpan("test.client")
	defer parentSpan.Finish()

	tracer.Inject(parentSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

	web.TracingMiddleware(tracer, "tracing-test")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(w, r)

	require.Len(t, tracer.FinishedSpans(), 1)
	span := tracer.FinishedSpans()[0]

	// Check if the trace id header was set.
	tid, err := jaeger.TraceIDFromString(w.Header().Get("X-Trace-Id"))
	require.NoError(t, err)

	assert.NotZero(t, tid)
	assert.EqualValues(t, span.SpanContext.TraceID, tid.Low)
}

func TestTracedTransport(t *testing.T) {
	var (
		method      = "PUT"
		tracer      = mocktracer.New()
		wireContext opentracing.SpanContext
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		wireContext, err = tracer.Extract(
			opentracing.HTTPHeaders,
			carrier,
		)
		require.NoError(t, err)
		assert.NotNil(t, wireContext)
	}))

	defer server.Close()

	client := http.Client{
		Transport: web.TracedTransport(tracer, http.DefaultTransport, "test.client"),
	}

	r, err := http.NewRequest(method, server.URL, nil)
	require.NoError(t, err)

	_, err = client.Do(r)
	require.NoError(t, err)

	require.NotEmpty(t, tracer.FinishedSpans())
	span := tracer.FinishedSpans()[0]

	assert.Equal(t, span.Context(), wireContext)

	assert.Equal(t, ext.SpanKindEnum("client"), span.Tags()[string(ext.SpanKind)])
	assert.Equal(t, server.URL, span.Tags()[string(ext.HTTPUrl)])
	assert.Equal(t, method, span.Tags()[string(ext.HTTPMethod)])
}

func TestTraceableRequestFunc(t *testing.T) {
	var (
		method      = "PUT"
		url         = "/dummy"
		tracer      = mocktracer.New()
		wireContext opentracing.SpanContext
	)

	span := tracer.StartSpan("test.client").(*mocktracer.MockSpan)
	defer span.Finish()

	r, err := web.TraceableRequestFunc(tracer)(span, httptest.NewRequest(method, url, nil))
	require.NoError(t, err)

	wireContext, err = tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	require.NoError(t, err)
	assert.Equal(t, span.Context(), wireContext)

	assert.Equal(t, ext.SpanKindEnum("client"), span.Tags()[string(ext.SpanKind)])
	assert.Equal(t, url, span.Tags()[string(ext.HTTPUrl)])
	assert.Equal(t, method, span.Tags()[string(ext.HTTPMethod)])
}

func TestTraceableRequestFuncNilSpan(t *testing.T) {
	var tracer = mocktracer.New()

	r := httptest.NewRequest("GET", "/", nil)
	got, err := web.TraceableRequestFunc(tracer)(nil, r)
	require.NoError(t, err)
	assert.Equal(t, r, got)
}
