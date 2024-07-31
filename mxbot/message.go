package mxbot

import (
	"maunium.net/go/mautrix/id"
)

// Message - message to send to the room (should be provided by the user)
type Message struct {
	RoomID id.RoomID // room to send the message
	Text   string    // text of the message
}
