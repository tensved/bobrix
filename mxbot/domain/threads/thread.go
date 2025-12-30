package domain

import (
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type MessagesThread struct {
	RoomID   id.RoomID
	ParentID id.EventID
	Messages []*event.Event
}
