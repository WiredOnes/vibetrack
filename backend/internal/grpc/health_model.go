package grpc

import (
	"github.com/leshless/pet/cub/internal/model"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var healthStatusFromModel = map[model.HealthStatus]healthpb.HealthCheckResponse_ServingStatus{
	model.HealthStatusUnknown:    healthpb.HealthCheckResponse_UNKNOWN,
	model.HealthStatusServing:    healthpb.HealthCheckResponse_SERVING,
	model.HealthStatusNotServing: healthpb.HealthCheckResponse_NOT_SERVING,
}
