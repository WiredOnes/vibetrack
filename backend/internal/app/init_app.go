package app

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/benbjohnson/clock"
	"github.com/leshless/golibrary/graceful"
	"github.com/leshless/golibrary/interrupt"
	"github.com/leshless/golibrary/stupid"
	cubpb "github.com/leshless/pet/cub/api/grpc/v1"
	httpapi "github.com/leshless/pet/cub/api/http/v1"
	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/db"
	"github.com/leshless/pet/cub/internal/environment"
	"github.com/leshless/pet/cub/internal/grpc"
	"github.com/leshless/pet/cub/internal/http"
	"github.com/leshless/pet/cub/internal/job"
	"github.com/leshless/pet/cub/internal/logic/health"
	"github.com/leshless/pet/cub/internal/state"
	"github.com/leshless/pet/cub/internal/telemetry"
	"go.uber.org/dig"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func InitApp(primitives Primitives) (*App, error) {
	c := dig.New()

	// Primitives
	c.Provide(stupid.Reflect(primitives))
	c.Provide(stupid.Reflect(primitives.Clock), dig.As(new(clock.Clock)))
	c.Provide(stupid.Reflect(primitives.Interrupter), dig.As(new(interrupt.Interrupter)))
	c.Provide(stupid.Reflect(primitives.FS), dig.As(new(fs.FS)))

	// Dependencies
	c.Provide(NewDependencies)
	c.Provide(graceful.NewManager, dig.As(new(graceful.Registrator), new(graceful.Terminator)))
	c.Provide(config.InitHolder, dig.As(new(config.Holder)))
	c.Provide(environment.InitHolder, dig.As(new(environment.Holder)))
	c.Provide(telemetry.InitLogger, dig.As(new(telemetry.Logger)))
	c.Provide(telemetry.InitRegistry, dig.As(new(telemetry.Registry)))
	c.Provide(telemetry.NewTelemetry)

	// Clients
	c.Provide(NewClients)
	c.Provide(db.InitClient, dig.As(new(db.Client)))

	// Adapters
	c.Provide(NewAdapters)

	// DB Adapters
	c.Provide(db.NewQueries)
	c.Provide(db.NewPingAdapter, dig.As(new(db.PingAdapter)))
	c.Provide(db.NewTxAdapter, dig.As(new(db.TxAdapter)))

	// State Adapters
	c.Provide(state.NewHealthAdapter, dig.As(new(state.HealthAdapter)))

	// Actions
	c.Provide(NewActions)

	// Controllers
	c.Provide(NewControllers)
	c.Provide(health.NewController, dig.As(new(health.Controller)))

	// Handlers
	c.Provide(NewHandlers)

	// GRPC Handlers
	c.Provide(grpc.NewHealthHandler, dig.As(new(healthpb.HealthServer)))
	c.Provide(stupid.Reflect(cubpb.UnimplementedPingServer{}))
	c.Provide(grpc.NewPingHandler, dig.As(new(cubpb.PingServer)))

	// HTTP Handlers
	c.Provide(http.NewHandlers, dig.As(new(httpapi.StrictServerInterface)))
	c.Provide(http.NewHealthHandler)

	// Jobs
	c.Provide(job.NewUpdateHealthStatus, dig.As(new(job.UpdateHealthStatus)))

	// Ports
	c.Provide(NewPorts)
	c.Provide(grpc.InitPort, dig.As(new(grpc.Port)))
	c.Provide(http.InitPort, dig.As(new(http.Port)))
	c.Provide(job.InitPort, dig.As(new(job.Port)))

	// App
	c.Provide(NewApp)

	var app App
	err := c.Invoke(func(a App) {
		app = a
	})
	if err != nil {
		return nil, fmt.Errorf("resolving app from the DI container: %w", err)
	}

	app.Logger.Info(context.Background(), "app successfully initialized")

	go func() {
		<-app.Interrupter.Context().Done()

		app.Logger.Warn(context.Background(), "graceful shutdown initiated")

		// Pass context.Background() here, since total timeout is passed trough config
		err := app.GracefulTerminator.Terminate(context.Background())
		if err != nil {
			app.Logger.Error(context.Background(), "app terminated with error", telemetry.Error(err))
		}
	}()

	return &app, nil
}
