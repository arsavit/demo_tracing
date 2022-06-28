package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type ctxTracer struct{}

func ContextWithTracer(ctx context.Context, tr trace.Tracer) context.Context {
	return context.WithValue(ctx, ctxTracer{}, tr)
}

func TracerFromContext(ctx context.Context) trace.Tracer {
	tr, ok := ctx.Value(ctxTracer{}).(trace.Tracer)
	if !ok {
		return nil
	}
	return tr
}
