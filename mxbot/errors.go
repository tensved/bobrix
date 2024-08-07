package mxbot

import "errors"

var (
	ErrNilMessage  = errors.New("message is nil")
	ErrSendMessage = errors.New("failed to send message")
)
