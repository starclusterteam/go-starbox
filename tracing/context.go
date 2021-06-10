package tracing

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
)

type contextSpanKey struct{}

// ContextWithSpanContext returns a context that holds a reference to the given span context.
func ContextWithSpanContext(ctx context.Context, spanContext opentracing.SpanContext) context.Context {
	return context.WithValue(ctx, contextSpanKey{}, spanContext)
}

// SpanContextFromContext extracts a span context from the given context.
// The returned span context may be nil if there is no span context stored in the given context.
func SpanContextFromContext(ctx context.Context) opentracing.SpanContext {
	return ctx.Value(contextSpanKey{}).(opentracing.SpanContext)
}
