package grpc

import (
	"context"

	"github.com/leshless/pet/cub/internal/logic/health"
	"github.com/leshless/pet/cub/internal/model"
	"github.com/leshless/pet/cub/internal/telemetry"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// @PublicPointerInstance
type healthHandler struct {
	telemetry.Telemetry
	controller health.Controller
}

var _ healthpb.HealthServer = (*healthHandler)(nil)

func (h *healthHandler) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	arg := checkRequestToDTO(req)

	res, err := h.controller.Check(ctx, arg)
	if err != nil {
		return nil, err
	}

	return checkResponseFromDTO(res), nil
}

func (h *healthHandler) List(ctx context.Context, req *healthpb.HealthListRequest) (*healthpb.HealthListResponse, error) {
	return nil, model.NewInternalError()
}

func (h *healthHandler) Watch(req *healthpb.HealthCheckRequest, srv grpc.ServerStreamingServer[healthpb.HealthCheckResponse]) error {
	return model.NewUnimplementedError()
}
