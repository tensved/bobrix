package ctx

import (
	"context"

	dctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"maunium.net/go/mautrix/event"
)

type Factory interface {
	New(ctx context.Context, evt *event.Event) (dctx.Ctx, error)
}
