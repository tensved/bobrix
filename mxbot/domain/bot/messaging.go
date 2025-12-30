package bot // ok

import (
	"context"

	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/id"
)

type BotMessaging interface {
	SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error
	JoinRoom(ctx context.Context, roomID id.RoomID) error
	Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error)

	StartTyping(ctx context.Context, roomID id.RoomID) error
	StopTyping(ctx context.Context, roomID id.RoomID) error

	Ping(ctx context.Context) error
}

