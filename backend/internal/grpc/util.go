package grpc

import (
	"github.com/leshless/pet/cub/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errorCodeFromModel = map[model.ErrorCode]codes.Code{
	model.ErrorCodeOK:                 codes.OK,
	model.ErrorCodeUnknown:            codes.Unknown,
	model.ErrorCodeBadRequest:         codes.InvalidArgument,
	model.ErrorCodeUnauthenticated:    codes.Unauthenticated,
	model.ErrorCodePermissionDenied:   codes.PermissionDenied,
	model.ErrorCodeNotFound:           codes.NotFound,
	model.ErrorCodeAlreadyExists:      codes.AlreadyExists,
	model.ErrorCodePreconditionFailed: codes.FailedPrecondition,
	model.ErrorCodeResourceExhausted:  codes.ResourceExhausted,
	model.ErrorCodeTimeoutExceeded:    codes.DeadlineExceeded,
	model.ErrorCodeInternal:           codes.Internal,
	model.ErrorCodeUnimplemented:      codes.Unimplemented,
}

func errorFromModel(err error) error {
	if err == nil {
		return nil
	}

	modelError, ok := err.(model.Error)
	if !ok {
		modelError = model.NewUnknownError()
	}

	grpcCode, ok := errorCodeFromModel[modelError.Code()]
	if !ok {
		grpcCode = codes.Unknown
	}

	return status.Error(grpcCode, modelError.Message())
}
