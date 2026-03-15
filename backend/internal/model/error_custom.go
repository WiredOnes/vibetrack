package model

// unauthenticatedError represents an error caused by missing or invalid credentials.
type unauthenticatedError struct{}

// NewUnauthenticatedError creates new unauthenticatedError.
func NewUnauthenticatedError() Error {
	return unauthenticatedError{}
}

var _ Error = unauthenticatedError{}

func (e unauthenticatedError) Error() string {
	return "unauthenticated"
}

func (e unauthenticatedError) Message() string {
	return "Unauthenticated"
}

func (e unauthenticatedError) Code() ErrorCode {
	return ErrorCodeUnauthenticated
}

// badRequestError represents an error caused by invalid request parameters.
type badRequestError struct{}

// NewBadRequestError creates new badRequestError.
func NewBadRequestError() Error {
	return badRequestError{}
}

var _ Error = badRequestError{}

func (e badRequestError) Error() string {
	return "bad_request"
}

func (e badRequestError) Message() string {
	return "Bad request"
}

func (e badRequestError) Code() ErrorCode {
	return ErrorCodeBadRequest
}
