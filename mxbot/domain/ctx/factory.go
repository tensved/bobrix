package ctx

import (
	"context"

	"maunium.net/go/mautrix/event"
)

// CtxFactory - creates domain context from event
type CtxFactory interface {
	New(ctx context.Context, evt *event.Event) (Ctx, error)
}
