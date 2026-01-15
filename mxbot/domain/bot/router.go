package bot

import (
	"context"

	"maunium.net/go/mautrix/event"
)

type EventRouter interface {
	HandleMatrixEvent(ctx context.Context, evt *event.Event) error
}

// type EventRouter interface {
// 	Dispatch(ctx context.Context, evt *event.Event) error
// }
