package app

import (
	"context"
	"fmt"
	"io/fs"

	httpapi "github.com/WiredOnes/vibetrack/backend/api/http/v1"
	"github.com/WiredOnes/vibetrack/backend/internal/config"
	"github.com/WiredOnes/vibetrack/backend/internal/db"
	"github.com/WiredOnes/vibetrack/backend/internal/environment"
	"github.com/WiredOnes/vibetrack/backend/internal/http"
	"github.com/WiredOnes/vibetrack/backend/internal/logic"
	"github.com/WiredOnes/vibetrack/backend/internal/state"
	"github.com/WiredOnes/vibetrack/backend/internal/telemetry"
	"github.com/benbjohnson/clock"
	"github.com/leshless/golibrary/graceful"
	"github.com/leshless/golibrary/interrupt"
	"github.com/leshless/golibrary/stupid"
	"go.uber.org/dig"
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
	c.Provide(logic.NewController, dig.As(new(logic.Controller)))

	// Handlers
	c.Provide(NewHandlers)

	// HTTP Handlers
	c.Provide(http.NewHandler, dig.As(new(httpapi.StrictServerInterface)))

	// Ports
	c.Provide(NewPorts)
	c.Provide(http.InitPort, dig.As(new(http.Port)))

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
