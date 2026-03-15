package http

import (
	"net/http"

	api "github.com/leshless/pet/cub/api/http/v1"
	"github.com/leshless/pet/cub/internal/model"
)

var statusFromModel = map[model.ErrorCode]int{
	model.ErrorCodeOK:                 http.StatusOK,
	model.ErrorCodeUnknown:            http.StatusInternalServerError,
	model.ErrorCodeBadRequest:         http.StatusBadRequest,
	model.ErrorCodeUnauthenticated:    http.StatusUnauthorized,
	model.ErrorCodePermissionDenied:   http.StatusForbidden,
	model.ErrorCodeNotFound:           http.StatusNotFound,
	model.ErrorCodeAlreadyExists:      http.StatusPreconditionFailed,
	model.ErrorCodePreconditionFailed: http.StatusPreconditionFailed,
	model.ErrorCodeResourceExhausted:  http.StatusPreconditionFailed,
	model.ErrorCodeTimeoutExceeded:    http.StatusRequestTimeout,
	model.ErrorCodeInternal:           http.StatusInternalServerError,
	model.ErrorCodeUnimplemented:      http.StatusNotImplemented,
}

func errorFromModel(err error) api.Error {
	modelErr, ok := err.(model.Error)
	if !ok {
		return api.Error{
			Message: model.NewUnknownError().Message(),
		}
	}

	return api.Error{
		Message: modelErr.Message(),
	}
}

func statusCodeFromModel(err error) int {
	modelErr, ok := err.(model.Error)
	if !ok {
		return statusFromModel[model.ErrorCodeUnknown]
	}

	status, ok := statusFromModel[modelErr.Code()]
	if !ok {
		return statusFromModel[model.ErrorCodeUnknown]
	}

	return status
}
