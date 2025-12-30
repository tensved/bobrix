package handlers

import (
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix/event"
)

// EventHandler - event handler for the bot
type EventHandler interface {
	EventType() event.Type     // provides access to the event type
	Filters() []filters.Filter // provides access to the event filters
	Handle(ctx.Ctx) error      // provides access to the event handler
}
