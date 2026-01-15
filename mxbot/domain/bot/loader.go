package bot

import (
	"context"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type EventLoader interface {
	GetEvent(ctx context.Context, roomID id.RoomID, eventID id.EventID) (*event.Event, error)
	// HandleMatrixEvent(ctx context.Context, evt *event.Event) error
}
