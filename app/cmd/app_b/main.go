package main

import (
	"context"
	"demo_tracing/app/handler"
	"demo_tracing/app/tracing"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Prepare B to start...")
	tr := tracing.MustGetTracer(os.Getenv("JAEGER_ENDPOINT"), "application B")
	ctx := tracing.ContextWithTracer(context.Background(), tr)

	r := chi.NewRouter()
	r.Use(tracing.RequestIDMiddleware)
	r.Use(tracing.TraceMiddleware(ctx))
	r.Get("/bye", handler.Bye)

	server := &http.Server{Addr: ":8092", Handler: r}

	log.Info().Msg("server B has been started")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Send()
	}
}
