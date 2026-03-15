package grpc

import (
	"context"

	cubpb "github.com/leshless/pet/cub/api/grpc/v1"
	"github.com/leshless/pet/cub/internal/telemetry"
)

// @PublicPointerInstance
type pingHandler struct {
	cubpb.UnimplementedPingServer
	telemetry.Telemetry
}

var _ cubpb.PingServer = (*pingHandler)(nil)

func (h *pingHandler) Ping(context.Context, *cubpb.PingRequest) (*cubpb.PingResponse, error) {
	return &cubpb.PingResponse{}, nil
}
