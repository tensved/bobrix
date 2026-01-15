package dispatcher

import (
	"github.com/tensved/bobrix/mxbot/application/ctx"
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
)

var _ bot.EventRouter = (*Dispatcher)(nil)

type Dispatcher struct {
	bot      bot.FullBot
	factory  ctx.Factory
	handlers []handlers.EventHandler
	filters  []filters.Filter
}

func New(
	bot bot.FullBot,
	factory ctx.Factory,
	handlers []handlers.EventHandler,
	globalFilters []filters.Filter,
) *Dispatcher {
	return &Dispatcher{
		bot:      bot,
		factory:  factory,
		handlers: handlers,
		filters:  globalFilters,
	}
}
