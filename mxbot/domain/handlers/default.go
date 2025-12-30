package handlers

import (
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	// dh "github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

var _ EventHandler = (*DefaultEventHandler)(nil)

type DefaultEventHandler struct {
	eventType event.Type
	filters   []filters.Filter
	handler   func(ctx.Ctx) error
}

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

func NewMessageHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return NewEventHandler(event.EventMessage, handler, filters...)
}

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
