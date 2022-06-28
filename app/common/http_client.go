package common

import (
	"demo_tracing/app/tracing"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/rs/zerolog/log"
)

type HTTPClient struct {
	http.Client
}

func NewClient() *HTTPClient {
	return &HTTPClient{
		Client: http.Client{
			Timeout:   time.Duration(60) * time.Second, //nolint:gomnd,gocritic
			Transport: LoggingRoundTripper{http.DefaultTransport},
		},
	}
}

func (c *HTTPClient) Process(request *http.Request) (body []byte, statusCode int, err error) {
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(middleware.RequestIDHeader, middleware.GetReqID(request.Context()))
	response, err := c.Do(request)

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Ctx(request.Context()).Error().Caller().Err(err).Msg("can't close body")
		}
	}()

	if err != nil {
		return nil, 0, err
	}

	body, err = io.ReadAll(response.Body)

	if err != nil {
		return nil, 0, err
	}

	return body, response.StatusCode, nil
}

type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var err error
	tr := tracing.TracerFromContext(r.Context())
	if tr == nil {
		response, err := lrt.Proxied.RoundTrip(r)
		if err != nil {
			return nil, err
		}
		return response, nil
	}

	_, span := tr.Start(r.Context(), "http.request "+r.URL.String())
	defer span.End()
	r.Header.Add(tracing.TraceIDHeader, span.SpanContext().TraceID().String())
	r.Header.Add(tracing.SpanIDHeader, span.SpanContext().SpanID().String())

	// Do
	response, err := lrt.Proxied.RoundTrip(r)
	if err != nil {
		log.Error().Err(err).Msg("error during request to external service")
		return nil, err
	}
	span.SetAttributes(
		attribute.String("request-id", middleware.GetReqID(r.Context())),
		attribute.String("request-url", r.URL.String()),
		attribute.Int("statusCode", response.StatusCode),
	)

	return response, nil
}
