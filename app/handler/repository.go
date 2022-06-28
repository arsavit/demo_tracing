package handler

import (
	"context"
	"demo_tracing/app/tracing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Save(ctx context.Context) {
	var span trace.Span
	tr := tracing.TracerFromContext(ctx)
	if tr != nil {
		_, span = tr.Start(ctx, "repository.Save")
		defer span.End()
	}
	log.Info().Msg("Saving model")
	if tr != nil {
		span.SetAttributes(
			attribute.String("request-id", middleware.GetReqID(ctx)),
			attribute.String("user-id", "123"),
		)
	}
}
