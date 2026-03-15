package http

import (
	"context"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	api "github.com/leshless/pet/cub/api/http/v1"
	"github.com/leshless/pet/cub/internal/telemetry"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

func recoveryMiddleware(tel telemetry.Telemetry) api.StrictMiddlewareFunc {
	return func(handler nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (response any, err error) {
			defer func() {
				if r := recover(); r != nil {
					tel.Logger.Error(ctx, "caught panic in http request", telemetry.Any("panic_value", r))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			return handler(ctx, w, r, req)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func telemetryMiddleware(tel telemetry.Telemetry, clock clock.Clock) api.StrictMiddlewareFunc {
	return func(handler nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (any, error) {
			ts := clock.Now()

			ctx = telemetry.ContextWith(ctx, telemetry.Method(r.Method), telemetry.Endpoint(r.URL.Path))

			tel.Logger.Info(ctx, "recieved http request")

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			res, err := handler(ctx, rw, r, req)

			latency := time.Since(ts)

			ctx = telemetry.ContextWith(
				ctx,
				telemetry.StatusCode(rw.statusCode),
				telemetry.Any("latency", latency),
			)

			tel.Registry.Counter(ctx, telemetry.HTTPRequestsTotal).Inc()
			tel.Registry.Summary(ctx, telemetry.HTTPRequestDurationSeconds).Observe(latency.Seconds())

			if err != nil || rw.statusCode != http.StatusOK {
				tel.Logger.Warn(
					ctx,
					"processed http request with error",
					telemetry.Error(err),
				)
			} else {
				tel.Logger.Info(ctx, "successfully processed http request")
			}

			return res, err
		}
	}
}
