package mxbot

import (
	// applctx "github.com/tensved/bobrix/mxbot/application/ctx"

	// "context"
	// "errors"

	"errors"

	applbot "github.com/tensved/bobrix/mxbot/application/bot"
	"maunium.net/go/mautrix/event"

	// applhandlers "github.com/tensved/bobrix/mxbot/application/handlers"
	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"

	// "github.com/tensved/bobrix/mxbot/domain/ctx"
	// domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	// "github.com/tensved/bobrix/mxbot/domain/filters"
	// "github.com/tensved/bobrix/mxbot/domain/handlers"
	"github.com/tensved/bobrix/mxbot/domain/threads"
	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	infrabot "github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
	// "maunium.net/go/mautrix/event"
)

// // facade
// // type DomCtx = domctx.Ctx
// // type DCtx = applctx.DefaultCtx

// // ---- main facades ----
// // type Bot = dombot.FullBot
// type Filter = filters.Filter
type EventHandler = handlers.EventHandler
type MessagesThread = threads.MessagesThread
type BotMedia = dombot.BotMedia

type Config = infrabot.Config

type BotCredentials = infracfg.BotCredentials

// type Ctx interface {
// 	domctx.Ctx
// 	Bot() dombot.BotInfo
// }

func NewMatrixBot(cfg Config) (*applbot.DefaultBot, error) {
	facade, err := infrabot.NewMatrixBot(cfg, []filters.Filter{})
	if err != nil {
		return nil, errors.New("err create facade")
	}

	return applbot.NewDefaultBot(
		cfg.Credentials.Username,
		facade,
		cfg.Logger,
		cfg.Credentials,
	), nil
}

// --- публичные handlers
// func AutoJoinRoomHandler(bot Bot) EventHandler {
// 	return applhandlers.AutoJoinRoomHandler(bot, bot)
// }

// func LoggerHandler(name string) EventHandler {
// 	return applhandlers.LoggerHandler()
// }

// // ---- handler factory ----

// // func NewEventHandler(
// // 	eventType event.Type,
// // 	fn func(Ctx) error,
// // ) EventHandler {
// // 	return handlers.NewEventHandler(eventType, fn)
// // }

// // func NewEventHandler(
// // 	eventType event.Type,
// // 	fn func(Ctx) error, // твой расширенный контекст с Bot()
// // ) EventHandler {
// // 	wrapped := func(c domctx.Ctx) error {
// // 		// пробуем привести к твоему расширенному интерфейсу
// // 		fullCtx, ok := c.(Ctx)
// // 		if !ok {
// // 			// если невозможно, делаем минимальную заглушку
// // 			return fn(&defaultCtxWrapper{Ctx: c})
// // 		}
// // 		return fn(fullCtx)
// // 	}

// // 	return handlers.NewEventHandler(eventType, wrapped)
// // }

// type defaultCtxWrapper struct {
// 	domctx.Ctx
// }

// func (d *defaultCtxWrapper) Bot() dombot.BotInfo {
// 	return nil // заглушка, если нет реального бота
// }

func NewEventHandler(
	eventType event.Type,
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewEventHandler(eventType, handler, filters...)
}

func NewMessageHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewMessageHandler(handler, filters...)
}

func NewStateMemberHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewStateMemberHandler(handler, filters...)
}
