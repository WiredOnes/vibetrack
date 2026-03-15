package http

import (
	api "github.com/WiredOnes/vibetrack/backend/api/http/v1"
	"github.com/WiredOnes/vibetrack/backend/internal/logic"
	"github.com/WiredOnes/vibetrack/backend/internal/model"
)

var healthStatusFromModel = map[model.HealthStatus]api.HealthStatus{
	model.HealthStatusUnknown:    api.UNKNOWN,
	model.HealthStatusServing:    api.SERVING,
	model.HealthStatusNotServing: api.NOTSERVING,
}

func checkHealthRequestToDTO(req api.CheckHealthRequestObject) logic.CheckArg {
	return logic.NewCheckArg()
}

func checkHealthResponseFromDTO(res logic.CheckRes) api.CheckHealth200JSONResponse {
	status, ok := healthStatusFromModel[res.Status]
	if !ok {
		status = api.UNKNOWN
	}

	return api.CheckHealth200JSONResponse{
		Status: status,
	}
}

func repositoryFromLogic(r logic.Repository) api.Repository {
	return api.Repository{
		Id:            int(r.ID),
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		Private:       r.Private,
		DefaultBranch: r.DefaultBranch,
		UpdatedAt:     r.UpdatedAt,
	}
}
