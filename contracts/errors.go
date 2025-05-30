package contracts

import "errors"

var (
	ErrMethodNotFound        = errors.New("method not found")
	ErrHandlerNotFound       = errors.New("handler not found")
	ErrInputRequired         = errors.New("input is required")
	ErrNoHealthCheckProvided = errors.New("no healthcheck provided")

	ErrCodeServiceNotFound = "service not found"
	ErrCodeMethodNotFound  = "method not found"
)
