package http

import (
	api "github.com/WiredOnes/vibetrack/backend/api/http/v1"
	"github.com/WiredOnes/vibetrack/backend/internal/logic/health"
	"github.com/WiredOnes/vibetrack/backend/internal/model"
)

var healthStatusFromModel = map[model.HealthStatus]api.HealthStatus{
	model.HealthStatusUnknown:    api.UNKNOWN,
	model.HealthStatusServing:    api.SERVING,
	model.HealthStatusNotServing: api.NOTSERVING,
}

func checkHealthRequestToDTO(req api.CheckHealthRequestObject) health.CheckArg {
	return health.NewCheckArg()
}

func checkHealthResponseFromDTO(res health.CheckRes) api.CheckHealth200JSONResponse {
	status, ok := healthStatusFromModel[res.Status]
	if !ok {
		status = api.UNKNOWN
	}

	return api.CheckHealth200JSONResponse{
		Status: status,
	}
}
