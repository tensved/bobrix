package contracts

import "errors"

var (
	ErrMethodNotFound        = errors.New("method not found")
	ErrHandlerNotFound       = errors.New("handler not found")
	ErrInputRequired         = errors.New("input is required")
	ErrNoHealthCheckProvided = errors.New("no healthcheck provided")
)

const (
	ErrCodeBadRequest           = 400 // invalid request / validation error
	ErrCodeServiceNotFound      = 404 // service not found
	ErrCodeMethodNotFound       = 405 // method not found
	ErrCodeInternalServiceError = 500 // internal server error
)
