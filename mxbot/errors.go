package mxbot

import "errors"

var (
	ErrNilMessage  = errors.New("message is nil")
	ErrNilEvent    = errors.New("event is nil")
	ErrSendMessage = errors.New("failed to send message")
	ErrUploadMedia = errors.New("failed to upload media file")
	ErrJoinToRoom  = errors.New("failed to join room")
)
