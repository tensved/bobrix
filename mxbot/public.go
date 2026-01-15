package mxbot

import (
	// applctx "github.com/tensved/bobrix/mxbot/application/ctx"
	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"github.com/tensved/bobrix/mxbot/domain/threads"
	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	infrabot "github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
	"maunium.net/go/mautrix/event"
)

// facade
// type DomCtx = domctx.Ctx
// type DCtx = applctx.DefaultCtx

// ---- main facades ----
type Bot = dombot.FullBot
type Filter = filters.Filter
type EventHandler = handlers.EventHandler
type MessagesThread = threads.MessagesThread
type BotMedia = dombot.BotMedia

type Config = infrabot.Config
type BotCredentials = infracfg.BotCredentials

type Ctx interface {
	domctx.Ctx
	Bot() dombot.BotInfo
}

func NewMatrixBot(cfg Config) (*infrabot.MatrixBot, error) {
	return infrabot.NewMatrixBot(cfg)
}

// ---- handler factory ----

// func NewEventHandler(
// 	eventType event.Type,
// 	fn func(Ctx) error,
// ) EventHandler {
// 	return handlers.NewEventHandler(eventType, fn)
// }

func NewEventHandler(
	eventType event.Type,
	fn func(Ctx) error, // твой расширенный контекст с Bot()
) EventHandler {
	wrapped := func(c domctx.Ctx) error {
		// пробуем привести к твоему расширенному интерфейсу
		fullCtx, ok := c.(Ctx)
		if !ok {
			// если невозможно, делаем минимальную заглушку
			return fn(&defaultCtxWrapper{Ctx: c})
		}
		return fn(fullCtx)
	}

	return handlers.NewEventHandler(eventType, wrapped)
}

type defaultCtxWrapper struct {
	domctx.Ctx
}

func (d *defaultCtxWrapper) Bot() dombot.BotInfo {
	return nil // заглушка, если нет реального бота
}
