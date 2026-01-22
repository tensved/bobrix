package handlers

import (
	"log/slog"

	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
)

var _ handlers.EventHandler = (*LoggerHandler)(nil)

type LoggerHandler struct {
	log *slog.Logger
}

func NewLoggerHandler(botName string, log ...*slog.Logger) *LoggerHandler {
	l := slog.Default()
	if len(log) > 0 {
		l = log[0]
	}
	return &LoggerHandler{log: l.With("bot", botName)}
}

func (h *LoggerHandler) EventType() event.Type {
	return event.EventMessage
}

// Filters returns nil — logger handles all events of its type
func (h *LoggerHandler) Filters() []filters.Filter {
	return nil
}

func (h *LoggerHandler) Handle(c ctx.Ctx) error {
	evt := c.Event()
	h.log.Info(
		"new event",
		"id", evt.ID,
		"sender", evt.Sender,
		"type", evt.Type,
		"content", evt.Content.Raw,
	)
	return nil
}
