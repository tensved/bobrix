package mxbot

import (
	"log/slog"

	applhandlers "github.com/tensved/bobrix/mxbot/application/handlers"
	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

func NewEventHandler(
	eventType event.Type,
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	wrapped := func(c domctx.Ctx) error {
		fc, ok := c.(Ctx)
		if !ok {
			return handler(&ctxWrapper{Ctx: c})
		}
		return handler(fc)
	}
	return handlers.NewEventHandler(eventType, wrapped, filters...)
}

// func NewStateMemberHandler2(
// 	handler func(ctx.Ctx) error,
// 	filters ...filters.Filter,
// ) EventHandler {
// 	return handlers.NewStateMemberHandler(handler, filters...)
// }

func NewStateMemberHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	wrapped := func(c domctx.Ctx) error {
		fc, ok := c.(Ctx)
		if !ok {
			return handler(&ctxWrapper{Ctx: c})
		}
		return handler(fc)
	}
	return handlers.NewStateMemberHandler(wrapped, filters...)
}

func NewMessageHandler(
	handler func(Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	wrapped := func(c domctx.Ctx) error {
		fc, ok := c.(Ctx)
		if !ok {
			return handler(&ctxWrapper{Ctx: c})
		}
		return handler(fc)
	}
	return handlers.NewMessageHandler(wrapped, filters...)
}

type ctxWrapper struct {
	domctx.Ctx
}

func (w *ctxWrapper) Bot() dombot.BotInfo {
	return nil
}

func AutoJoinRoomHandler(bot Bot) EventHandler {
	return applhandlers.AutoJoinRoomHandler(bot, bot)
}

func NewLoggerHandler(name string, log ...*slog.Logger) EventHandler {
	return applhandlers.NewLoggerHandler(name, log...)
}
