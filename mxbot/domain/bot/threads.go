package bot

import (
	"context"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/tensved/bobrix/mxbot/domain/threads"
)

type BotThreads interface {
	IsThreadEnabled() bool
	GetThreadByEvent(ctx context.Context, evt *event.Event) (*threads.MessagesThread, error)
	GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*threads.MessagesThread, error)
}
