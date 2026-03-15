package grpc

import (
	"github.com/leshless/pet/cub/internal/logic/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func checkRequestToDTO(req *healthpb.HealthCheckRequest) health.CheckArg {
	return health.NewCheckArg()
}

func checkResponseFromDTO(res health.CheckRes) *healthpb.HealthCheckResponse {
	status, ok := healthStatusFromModel[res.Status]
	if !ok {
		status = healthpb.HealthCheckResponse_UNKNOWN
	}

	return &healthpb.HealthCheckResponse{
		Status: status,
	}
}
