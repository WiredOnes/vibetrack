package http

import (
	"context"

	api "github.com/WiredOnes/vibetrack/backend/api/http/v1"
	"github.com/WiredOnes/vibetrack/backend/internal/logic"
	"github.com/WiredOnes/vibetrack/backend/internal/model"
	"github.com/WiredOnes/vibetrack/backend/internal/telemetry"
)

// @PublicPointerInstance
type Handler struct {
	telemetry.Telemetry
	controller logic.Controller
}

var _ api.StrictServerInterface = (*Handler)(nil)

func (h *Handler) CheckHealth(ctx context.Context, req api.CheckHealthRequestObject) (api.CheckHealthResponseObject, error) {
	arg := checkHealthRequestToDTO(req)

	res, err := h.controller.CheckHealth(ctx, arg)
	if err != nil {
		return api.CheckHealthdefaultJSONResponse{
			Body:       errorFromModel(err),
			StatusCode: statusCodeFromModel(err),
		}, nil
	}

	return checkHealthResponseFromDTO(res), nil
}

func (h *Handler) GetRepositories(ctx context.Context, req api.GetRepositoriesRequestObject) (api.GetRepositoriesResponseObject, error) {
	token := bearerTokenFromContext(ctx)
	if token == "" {
		return api.GetRepositoriesdefaultJSONResponse{
			Body:       errorFromModel(model.NewUnauthenticatedError()),
			StatusCode: statusCodeFromModel(model.NewUnauthenticatedError()),
		}, nil
	}

	res, err := h.controller.GetRepositories(ctx, logic.GetRepositoriesArg{Token: token})
	if err != nil {
		return api.GetRepositoriesdefaultJSONResponse{
			Body:       errorFromModel(err),
			StatusCode: statusCodeFromModel(err),
		}, nil
	}

	repos := make(api.GetRepositories200JSONResponse, len(res.Repositories))
	for i, r := range res.Repositories {
		repos[i] = repositoryFromLogic(r)
	}

	return repos, nil
}

func (h *Handler) GetRepositoriesRepositoryID(ctx context.Context, req api.GetRepositoriesRepositoryIDRequestObject) (api.GetRepositoriesRepositoryIDResponseObject, error) {
	return api.GetRepositoriesRepositoryIDdefaultJSONResponse{
		Body:       errorFromModel(model.NewUnimplementedError()),
		StatusCode: statusCodeFromModel(model.NewUnimplementedError()),
	}, nil
}

func (h *Handler) PostRepositoryRepositoryIDAnalyze(ctx context.Context, req api.PostRepositoryRepositoryIDAnalyzeRequestObject) (api.PostRepositoryRepositoryIDAnalyzeResponseObject, error) {
	return api.PostRepositoryRepositoryIDAnalyzedefaultJSONResponse{
		Body:       errorFromModel(model.NewUnimplementedError()),
		StatusCode: statusCodeFromModel(model.NewUnimplementedError()),
	}, nil
}

func (h *Handler) PostRepositoryRepositoryIDCommitSHAAnalyze(ctx context.Context, req api.PostRepositoryRepositoryIDCommitSHAAnalyzeRequestObject) (api.PostRepositoryRepositoryIDCommitSHAAnalyzeResponseObject, error) {
	return api.PostRepositoryRepositoryIDCommitSHAAnalyzedefaultJSONResponse{
		Body:       errorFromModel(model.NewUnimplementedError()),
		StatusCode: statusCodeFromModel(model.NewUnimplementedError()),
	}, nil
}
