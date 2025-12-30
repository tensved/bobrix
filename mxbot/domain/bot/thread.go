package bot // ok

import (
	"context"

	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"maunium.net/go/mautrix/event"
)

type Threads interface {
	IsThreadEnabled() bool
	GetThreadByEvent(ctx context.Context, evt *event.Event) (*threads.MessagesThread, error)
}
