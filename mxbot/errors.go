package mxbot

import "errors"

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNilMessage     = errors.New("message is nil")
	ErrSendMessage    = errors.New("failed to send message")
)
