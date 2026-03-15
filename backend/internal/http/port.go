package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/WiredOnes/vibetrack/backend/internal/telemetry"
)

type Port interface {
	Run(ctx context.Context) error
}

// @PublicPointerInstance
type port struct {
	telemetry.Telemetry
	httpServer *http.Server
}

var _ Port = (*port)(nil)

func (p *port) Run(ctx context.Context) error {
	p.Logger.Info(ctx, "starting http server")

	err := p.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		p.Logger.Error(ctx, "http server stopped with error", telemetry.Error(err))
		return err
	}

	p.Logger.Warn(ctx, "http server stopped")

	return nil
}
