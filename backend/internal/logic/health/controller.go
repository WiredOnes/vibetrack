package health

import (
	"context"

	"github.com/leshless/pet/cub/internal/db"
	"github.com/leshless/pet/cub/internal/model"
	"github.com/leshless/pet/cub/internal/state"
	"github.com/leshless/pet/cub/internal/telemetry"
)

type Controller interface {
	Check(ctx context.Context, arg CheckArg) (CheckRes, error)
	UpdateStatus(ctx context.Context, arg UpdateStatusArg) (UpdateStatusRes, error)
}

// @PublicPointerInstance
type controller struct {
	telemetry.Telemetry
	healthState state.HealthAdapter
	pingDB      db.PingAdapter
}

var _ Controller = (*controller)(nil)

// @PublicValueInstance
type CheckArg struct{}

// @PublicValueInstance
type CheckRes struct {
	Status model.HealthStatus
}

func (c *controller) Check(ctx context.Context, arg CheckArg) (CheckRes, error) {
	status, err := c.healthState.GetStatus(ctx)
	if err != nil {
		c.Logger.Error(ctx, "failed to get health status state", telemetry.Error(err))
		return CheckRes{}, err
	}

	return NewCheckRes(status), nil
}

// @PublicValueInstance
type UpdateStatusArg struct{}

// @PublicValueInstance
type UpdateStatusRes struct{}

func (c *controller) UpdateStatus(ctx context.Context, arg UpdateStatusArg) (UpdateStatusRes, error) {
	status := model.HealthStatusServing

	err := c.pingDB.Ping(ctx)
	if err != nil {
		c.Logger.Error(ctx, "failed to ping database", telemetry.Error(err))
		status = model.HealthStatusNotServing
	}

	err = c.healthState.SetStatus(ctx, status)
	if err != nil {
		c.Logger.Error(ctx, "failed to update health status state", telemetry.Error(err))
		return NewUpdateStatusRes(), err
	}

	return NewUpdateStatusRes(), nil
}
