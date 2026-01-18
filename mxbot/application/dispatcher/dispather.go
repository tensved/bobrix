package dispatcher

import (
	// "github.com/tensved/bobrix/mxbot/application/ctx"

	"github.com/rs/zerolog"
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
)

var _ bot.EventRouter = (*Dispatcher)(nil)

type Dispatcher struct {
	bot      bot.FullBot
	factory  ctx.CtxFactory
	handlers []handlers.EventHandler
	filters  []filters.Filter // busines logic
	// typingTimeout time.Duration
	logger *zerolog.Logger
}

func New(
	bot bot.FullBot,
	factory ctx.CtxFactory,
	handlers []handlers.EventHandler,
	globalFilters []filters.Filter,
	logger *zerolog.Logger,
	// typingTimeout time.Duration,
) *Dispatcher {
	return &Dispatcher{
		bot:      bot,
		factory:  factory,
		handlers: handlers,
		filters:  globalFilters,
		logger:   logger,
	}
}
