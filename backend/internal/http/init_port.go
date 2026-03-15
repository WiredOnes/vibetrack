package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/benbjohnson/clock"
	"github.com/go-chi/chi/v5"
	"github.com/leshless/golibrary/graceful"
	api "github.com/leshless/pet/cub/api/http/v1"
	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/environment"
	"github.com/leshless/pet/cub/internal/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitPort(
	logger telemetry.Logger,
	tel telemetry.Telemetry,
	clock clock.Clock,
	configHolder config.Holder,
	environmentHolder environment.Holder,
	gracefulRegistrator graceful.Registrator,
	handler api.StrictServerInterface,
) (*port, error) {
	ctx := context.Background()

	logger.Info(ctx, "initializing http server")

	middlewares := []api.StrictMiddlewareFunc{
		telemetryMiddleware(tel, clock),
		recoveryMiddleware(tel),
	}

	chiRouter := chi.NewRouter()
	api.HandlerFromMux(api.NewStrictHandler(handler, middlewares), chiRouter)

	chiRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	chiRouter.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	chiRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	config := configHolder.Config().HTTP
	environment := environmentHolder.Environment()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      chiRouter,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	if config.EnableTLS {
		cert, err := tls.X509KeyPair([]byte(environment.TLSCertificate), []byte(environment.TLSKey))
		if err != nil {
			logger.Error(ctx, "failed to create x509 key pair for HTTPS", telemetry.Error(err))
			return nil, fmt.Errorf("creating x509 key pair for HTTPS: %w", err)
		}

		httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			ClientAuth:   tls.NoClientCert,
		}
	}

	gracefulRegistrator.Register(httpServer.Shutdown)

	port := NewPort(
		tel,
		httpServer,
	)

	logger.Info(ctx, "http port successfully initialized")

	return port, nil
}
