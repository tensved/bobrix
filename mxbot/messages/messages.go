package messages

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// Message - message interface
// should be implemented by all message types
// It is used to send messages to the room by the bot
type Message interface {
	Type() event.MessageType                             // get message type
	AsEvent() event.MessageEventContent                  // get event content
	AsReqUpload(roomID id.RoomID) mautrix.ReqUploadMedia // get upload media request
	SetContentURI(contentURI id.ContentURI)              // set content URI after upload
}
