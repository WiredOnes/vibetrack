package model

type Error interface {
	error
	Message() string
	Code() ErrorCode
}

// @Enum
type ErrorCode uint8

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeOK
	ErrorCodeBadRequest
	ErrorCodeUnauthenticated
	ErrorCodePermissionDenied
	ErrorCodeNotFound
	ErrorCodeAlreadyExists
	ErrorCodePreconditionFailed
	ErrorCodeResourceExhausted
	ErrorCodeTimeoutExceeded
	ErrorCodeInternal
	ErrorCodeUnimplemented
)

// @Error
type unknownError struct{}

var (
	unknownErrorText    = "unknown"
	unknownErrorMessage = "Unknown"
	unknownErrorCode    = ErrorCodeUnknown
)

// @Error
type internalError struct{}

var (
	internalErrorText    = "internal"
	internalErrorMessage = "Internal"
	internalErrorCode    = ErrorCodeInternal
)

// @Error
type unimplementedError struct{}

var (
	unimplementedErrorText    = "unimplemented"
	unimplementedErrorMessage = "Unimplemented"
	unimplementedErrorCode    = ErrorCodeUnimplemented
)
