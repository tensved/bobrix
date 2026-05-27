package bot

import (
	"context"
	"time"

	"maunium.net/go/mautrix/id"
)

type BotTyping interface {
	// EnsureTyping starts (or extends) the typing loop and returns a stop func.
	// Calling the stop func signals that this caller is done; the loop stops
	// only when all callers have stopped (ref-counted) or the TTL fallback fires.
	EnsureTyping(ctx context.Context, roomID id.RoomID, ttl time.Duration) func()
	LoopTyping(ctx context.Context, roomID id.RoomID) (cancelTyping func(), done <-chan struct{}, err error)
	StartTyping(ctx context.Context, roomID id.RoomID) error
	StopTyping(ctx context.Context, roomID id.RoomID) error

	GetTypingTimeout() time.Duration
}
