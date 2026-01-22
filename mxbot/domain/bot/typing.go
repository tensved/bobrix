package bot

import (
	"context"

	"maunium.net/go/mautrix/id"
)

type BotTyping interface {
	LoopTyping(ctx context.Context, roomID id.RoomID) (cancelTyping func(), err error)
	StartTyping(ctx context.Context, roomID id.RoomID) error
	StopTyping(ctx context.Context, roomID id.RoomID) error
}
