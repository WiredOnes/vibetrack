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
