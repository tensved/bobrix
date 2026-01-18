package bot

import (
	"context"

	"maunium.net/go/mautrix/event"
)

// EventSink is an inbound port.
// It represents something that can receive Matrix events from the outside world
// (Matrix syncer, webhooks, tests, etc) and pass them into the application.
type EventSink interface {
	HandleMatrixEvent(ctx context.Context, evt *event.Event) error
	// AddHandler(handler EventHandler)
}

type Auth interface {
	Reauthorize(ctx context.Context) error
}
