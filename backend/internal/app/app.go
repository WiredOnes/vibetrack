package app

import (
	"context"
	"fmt"

	"github.com/leshless/pet/cub/internal/telemetry"
	"golang.org/x/sync/errgroup"
)

// @PublicValueInstance
type App struct {
	Primitives
	Dependencies
	Adapters
	Controllers
	Actions
	Handlers
	Ports
}

func (app *App) Run() error {
	app.Logger.Info(context.Background(), "starting app")

	ctx := app.Interrupter.Context()

	var eg errgroup.Group

	eg.Go(func() error {
		return app.GRPC.Run(ctx)
	})
	eg.Go(func() error {
		return app.HTTP.Run(ctx)
	})
	eg.Go(func() error {
		return app.Job.Run(ctx)
	})

	err := eg.Wait()
	if err != nil {
		app.Logger.Error(context.Background(), "app stopped with error", telemetry.Error(err))
		return fmt.Errorf("stopping app: %w", err)
	}

	app.Logger.Info(context.Background(), "app successfully stopped")

	return nil
}
