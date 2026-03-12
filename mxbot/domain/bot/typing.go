package bot

import (
	"context"
	"time"

	"maunium.net/go/mautrix/id"
)

type BotTyping interface {
	EnsureTyping(ctx context.Context, roomID id.RoomID, ttl time.Duration)
	LoopTyping(ctx context.Context, roomID id.RoomID) (cancelTyping func(), done <-chan struct{}, err error)
	StartTyping(ctx context.Context, roomID id.RoomID) error
	StopTyping(ctx context.Context, roomID id.RoomID) error

	GetTypingTimeout() time.Duration
}
