package mxbot

import (
	"log/slog"
	"maunium.net/go/mautrix/event"
)

type JoinRoomParams struct {
	PreJoinHook   func(ctx Ctx) error
	AfterJoinHook func(ctx Ctx) error
}

// AutoJoinRoomHandler - join the room on invite automatically
// You can pass JoinRoomParams to modify the behavior of the handler
// Use PreJoinHook to modify the behavior before joining the room
// If PreJoinHook returns an error, the join is aborted
// Use AfterJoinHook to modify the behavior after joining the room
func AutoJoinRoomHandler(bot Bot, params ...JoinRoomParams) EventHandler {
	return NewStateMemberHandler(func(ctx Ctx) error {
		evt := ctx.Event()

		p := JoinRoomParams{}

		if len(params) > 0 {
			p = params[0]
		}

		if p.PreJoinHook != nil {
			if err := p.PreJoinHook(ctx); err != nil {
				return err
			}
		}

		if err := bot.JoinRoom(ctx.Context(), evt.RoomID); err != nil {
			return err
		}

		if p.AfterJoinHook != nil {
			if err := p.AfterJoinHook(ctx); err != nil {
				return err
			}
		}

		return nil
	}, FilterInviteMe(bot))
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
