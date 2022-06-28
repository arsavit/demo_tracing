package tracing

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/oklog/ulid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	TraceIDHeader = "X-Trace-Id"
	SpanIDHeader  = "X-Span-Id"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(middleware.RequestIDHeader)
		if requestID == "" {
			requestID = ulid.MustNew(ulid.Now(), rand.Reader).String()
		}
		ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)
		w.Header().Set(middleware.RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func TraceMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tr := TracerFromContext(ctx)
			if tr == nil {
				log.Error().Msg("tracer from context is nil")
				next.ServeHTTP(w, r)
				return
			}

			newCtx := ContextWithTracer(r.Context(), tr)

			parentTraceID := r.Header.Get(TraceIDHeader)
			parentSpanID := r.Header.Get(SpanIDHeader)
			spanContext := newChildSpanContext(newCtx, parentTraceID, parentSpanID)
			newCtx = trace.ContextWithSpanContext(newCtx, spanContext)

			newCtx, span := tr.Start(newCtx, "Middleware")
			defer span.End()

			span.SetAttributes(attribute.String("request-id", middleware.GetReqID(newCtx)))

			w.Header().Set(TraceIDHeader, span.SpanContext().TraceID().String())
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

func newChildSpanContext(ctx context.Context, parentTraceID, parentSpanID string) trace.SpanContext {
	var emptySpanContext trace.SpanContext
	if parentTraceID == "" || parentSpanID == "" {
		return emptySpanContext
	}

	traceID, err := trace.TraceIDFromHex(parentTraceID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("cannot parse parent traceID, using empty spanContext")
		return emptySpanContext
	}

	spanID, err := trace.SpanIDFromHex(parentSpanID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("cannot parse parent spanID, using empty spanContext")
		return emptySpanContext
	}

	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 0o1
	spanContextConfig.Remote = true
	spanContext := trace.NewSpanContext(spanContextConfig)
	if !spanContext.IsValid() {
		return emptySpanContext
	}
	return spanContext
}
