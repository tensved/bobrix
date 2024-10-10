package messages_v2

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Message interface {
	Type() event.MessageType                                // get message type
	AsEvent(rel *event.RelatesTo) event.MessageEventContent // get event contentBytes
	AsJSON() map[string]any                                 // get message as JSON
	AsReqUpload() mautrix.ReqUploadMedia                    // get upload media request
	SetContentURI(contentURI id.ContentURI)                 // set contentBytes URI after upload

	AddCustomFields(values ...any) // add custom fields to message. Required use as < key, value, ... >
}

type BaseMessage struct {
	msgType event.MessageType

	text         string
	contentURI   id.ContentURI
	customFields map[string]any
}
