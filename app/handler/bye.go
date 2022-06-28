package handler

import (
	"demo_tracing/app/tracing"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func Bye(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := tracing.TracerFromContext(ctx)
	_, span := tr.Start(ctx, "in-handler")
	defer span.End()
	w.Header().Set("Content-Type", "application/json")

	requestID := middleware.GetReqID(ctx)
	resp := Response{
		RequestID: requestID,
		TraceID:   span.SpanContext().TraceID().String(),
		SpanID:    span.SpanContext().SpanID().String(),
		Message:   "Bye",
	}

	response, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Ctx(ctx).Error().Caller().Err(err).Msg("can't marshal response")
		response = make([]byte, 0)
	}

	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Ctx(ctx).Error().Err(err).Msg("can't write to response writer")
	}
}
