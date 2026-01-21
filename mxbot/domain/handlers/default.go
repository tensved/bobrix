package handlers

import (
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"

	"maunium.net/go/mautrix/event"
)

var _ EventHandler = (*DefaultEventHandler)(nil)

// DefaultEventHandler - default implementation of the EventHandler
type DefaultEventHandler struct {
	eventType event.Type
	filters   []filters.Filter
	handler   func(ctx.Ctx) error
}

// NewEventHandler - EventHandler constructor
// eventType - type of the event
// handler - handler of the event
// filters - filters of the event (optional)
func NewEventHandler(
	eventType event.Type,
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return &DefaultEventHandler{
		eventType: eventType,
		filters:   filters,
		handler:   handler,
	}
}

// NewMessageHandler - EventHandler constructor for message events
// It is a wrapper for NewEventHandler
func NewMessageHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return NewEventHandler(event.EventMessage, handler, filters...)
}

// NewStateMemberHandler - EventHandler constructor for state member events (join/leave room, etc...)
// It is a wrapper for NewEventHandler
func NewStateMemberHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return NewEventHandler(event.StateMember, handler, filters...)
}

func (h *DefaultEventHandler) EventType() event.Type {
	return h.eventType
}

func (h *DefaultEventHandler) Filters() []filters.Filter {
	return h.filters
}

func (h *DefaultEventHandler) Handle(c ctx.Ctx) error {
	for _, f := range h.filters {
		if !f(c.Event()) {
			return nil
		}
	}
	return h.handler(c)
}
