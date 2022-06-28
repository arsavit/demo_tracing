package handler

import (
	"demo_tracing/app/common"
	"demo_tracing/app/tracing"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Response struct {
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id"`
	SpanID    string `json:"span_id"`
	Message   string `json:"message"`
}

func Hello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := tracing.TracerFromContext(ctx)
	ctx, span := tr.Start(ctx, "in-handler")
	defer span.End()
	w.Header().Set("Content-Type", "application/json")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://app_b:8092/bye", http.NoBody)
	if err != nil {
		log.Ctx(ctx).Err(err).Send()
	}

	client := common.NewClient()

	_, _, err = client.Process(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Ctx(ctx).Error().Err(err).Msg("Can't process request to app_b")
		return
	}

	Save(ctx)

	requestID := middleware.GetReqID(ctx)
	resp := Response{
		RequestID: requestID,
		TraceID:   span.SpanContext().TraceID().String(),
		SpanID:    span.SpanContext().SpanID().String(),
		Message:   "Hello",
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
