package bot // ok

import (
	"context"

	"maunium.net/go/mautrix/event"
)

type Crypto interface {
	DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error)
	IsEncrypted(evt *event.Event) bool
}

type OlmMachine interface {
	DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error)
}
