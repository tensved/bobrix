package contracts

import "errors"

var (
	ErrMethodNotFound  = errors.New("method not found")
	ErrInvalidData     = errors.New("invalid data")
	ErrNotImplemented  = errors.New("not implemented")
	ErrHandlerNotFound = errors.New("handler not found")

	ErrInputRequired = errors.New("input is required")
)
