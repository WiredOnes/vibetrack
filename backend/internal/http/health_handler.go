package http

import (
	"context"

	api "github.com/leshless/pet/cub/api/http/v1"
	"github.com/leshless/pet/cub/internal/logic/health"
	"github.com/leshless/pet/cub/internal/telemetry"
)

// @PublicPointerInstance
type HealthHandler struct {
	telemetry.Telemetry
	controller health.Controller
}

func (h *HealthHandler) CheckHealth(ctx context.Context, req api.CheckHealthRequestObject) (api.CheckHealthResponseObject, error) {
	arg := checkHealthRequestToDTO(req)

	res, err := h.controller.Check(ctx, arg)
	if err != nil {
		return api.CheckHealthdefaultJSONResponse{
			Body:       errorFromModel(err),
			StatusCode: statusCodeFromModel(err),
		}, nil
	}

	return checkHealthResponseFromDTO(res), nil
}
