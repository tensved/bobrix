package mxbot

import (
	"context"
	"log/slog"
	"maunium.net/go/mautrix/event"
)

func AutoJoinRoomHandler(bot Bot) EventHandler {
	return NewStateMemberHandler(func(ctx Ctx) error {
		evt := ctx.Event()

		if evt.Content.AsMember().Membership == event.MembershipInvite {
			err := bot.JoinRoom(context.TODO(), evt.RoomID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

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
	sender := ctx.Event().Sender
	eventType := ctx.Event().Type
	content := ctx.Event().Content

	h.log.Info("new event", "sender", sender, "type", eventType, "content", content)
	return nil
}

func (h *LoggerHandler) EventType() event.Type {
	return event.EventMessage
}

func (h *LoggerHandler) Filters() []Filter {
	return []Filter{}
}
