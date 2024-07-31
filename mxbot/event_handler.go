package mxbot

import "maunium.net/go/mautrix/event"

// EventHandler - event handler for the bot
type EventHandler interface {
	EventType() event.Type // provides access to the event type
	Filters() []Filter     // provides access to the event filters
	Handle(ctx Ctx) error  // provides access to the event handler
}

// NewEventHandler - EventHandler constructor
// eventType - type of the event
// handler - handler of the event
// filters - filters of the event (optional)
func NewEventHandler(eventType event.Type, handler func(ctx Ctx) error, filters ...Filter) EventHandler {
	return &DefaultEventHandler{
		eventType: eventType,
		filters:   filters,
		handler:   handler,
	}
}

// NewMessageHandler - EventHandler constructor for message events
// It is a wrapper for NewEventHandler
func NewMessageHandler(handler func(ctx Ctx) error, filters ...Filter) EventHandler {
	return NewEventHandler(event.EventMessage, handler, filters...)
}

// NewStateMemberHandler - EventHandler constructor for state member events (join/leave room, etc...)
// It is a wrapper for NewEventHandler
func NewStateMemberHandler(handler func(ctx Ctx) error, filters ...Filter) EventHandler {
	return NewEventHandler(event.StateMember, handler, filters...)
}

var _ EventHandler = (*DefaultEventHandler)(nil)

// DefaultEventHandler - default implementation of the EventHandler
type DefaultEventHandler struct {
	eventType event.Type
	filters   []Filter
	handler   func(ctx Ctx) error
}

func (h *DefaultEventHandler) EventType() event.Type {
	return h.eventType
}
func (h *DefaultEventHandler) Filters() []Filter {
	return h.filters
}

func (h *DefaultEventHandler) Handle(ctx Ctx) error {
	for _, filter := range h.filters {
		if !filter(ctx.Event()) {
			return nil
		}
	}

	return h.handler(ctx)
}
