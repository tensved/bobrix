package handlers

import (
	"log/slog"

	"maunium.net/go/mautrix/event"
)

var _ EventHandler = (*LoggerHandler)(nil)

type LoggerHandler struct {
	log *slog.Logger
}

func NewLoggerHandler(botName string, log ...*slog.Logger) *LoggerHandler {
	logger := slog.Default()

	if len(log) > 0 {
		logger = log[0]
	}

	logger = logger.With("bot", botName)

	return &LoggerHandler{
		log: logger,
	}
}

func (h *LoggerHandler) Handle(ctx Ctx) error {
	evt := ctx.Event()

	id := evt.ID
	sender := evt.Sender
	eventType := evt.Type
	content := evt.Content

	h.log.Info("new event", "id", id, "sender", sender, "type", eventType, "content", content.Raw)
	return nil
}

func (h *LoggerHandler) EventType() event.Type {
	return event.EventMessage
}

func (h *LoggerHandler) Filters() []Filter {
	return []Filter{}
}
