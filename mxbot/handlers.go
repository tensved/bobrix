package mxbot

import (
	"log/slog"

	applhandlers "github.com/tensved/bobrix/mxbot/application/handlers"
	// dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

func NewEventHandler(
	eventType event.Type,
	handler func(Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewEventHandler(eventType, handler, filters...)
}

func NewStateMemberHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewStateMemberHandler(handler, filters...)
}

func NewMessageHandler(
	handler func(Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewMessageHandler(handler, filters...)
}

// func NewEventHandler(
// 	eventType event.Type,
// 	handler func(Ctx) error,
// 	filters ...filters.Filter,
// ) EventHandler {
// 	wrapped := func(c domctx.Ctx) error {
// 		fc, ok := c.(Ctx)
// 		if !ok {
// 			return handler(&ctxWrapper{Ctx: c})
// 		}
// 		return handler(fc)
// 	}
// 	return handlers.NewEventHandler(eventType, wrapped, filters...)
// }

// func NewStateMemberHandler(
// 	handler func(ctx.Ctx) error,
// 	filters ...filters.Filter,
// ) EventHandler {
// 	wrapped := func(c domctx.Ctx) error {
// 		fc, ok := c.(Ctx)
// 		if !ok {
// 			return handler(&ctxWrapper{Ctx: c})
// 		}
// 		return handler(fc)
// 	}
// 	return handlers.NewStateMemberHandler(wrapped, filters...)
// }

// func NewMessageHandler(
// 	handler func(Ctx) error,
// 	filters ...filters.Filter,
// ) EventHandler {
// 	wrapped := func(c domctx.Ctx) error {
// 		fc, ok := c.(Ctx)
// 		if !ok {
// 			return handler(&ctxWrapper{Ctx: c})
// 		}
// 		return handler(fc)
// 	}
// 	return handlers.NewMessageHandler(wrapped, filters...)
// }

// type ctxWrapper struct {
// 	domctx.Ctx
// }

// func (w *ctxWrapper) Bot() dombot.BotInfo {
// 	return nil
// }

func AutoJoinRoomHandler(bot Bot, params ...applhandlers.JoinRoomParams) EventHandler {
	return applhandlers.AutoJoinRoomHandler(bot, bot, params...)
}

func NewLoggerHandler(name string, log ...*slog.Logger) EventHandler {
	return applhandlers.NewLoggerHandler(name, log...)
}

type JoinRoomParams = applhandlers.JoinRoomParams

// func WrapJoinRoomParams(p JoinRoomParams) JoinRoomParams {
// 	return JoinRoomParams{
// 		PreJoinHook:   wrapCtx(p.PreJoinHook),
// 		AfterJoinHook: wrapCtx(p.AfterJoinHook),
// 	}
// }

// func wrapCtx(
// 	handler func(Ctx) error,
// ) func(ctx.Ctx) error {
// 	if handler == nil {
// 		return nil
// 	}

// 	return func(c ctx.Ctx) error {
// 		if mx, ok := c.(Ctx); ok {
// 			return handler(mx)
// 		}

// 		// fallback — безопасный wrapper
// 		return handler(&ctxWrapper{Ctx: c})
// 	}
// }

// type ctxWrapper struct {
// 	ctx.Ctx
// }

// func (w *ctxWrapper) Bot() dombot.BotInfo {
// 	// если реально нельзя — лучше panic, чем nil
// 	panic("Bot() is not available in this context")
// }
