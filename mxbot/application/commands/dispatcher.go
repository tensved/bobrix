package commands

import (
	applfilters "github.com/tensved/bobrix/mxbot/application/filters"
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/commands"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
)

type Dispatcher struct {
	handlers []handlers.EventHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: []handlers.EventHandler{},
	}
}

// AddCommand
func Register(
	dispatcher bot.EventDispatcher,
	cmd *commands.Command,
	extraFilters ...filters.Filter,
) {
	dispatcher.AddEventHandler(
		handlers.NewMessageHandler(
			cmd.Handler,
			append(extraFilters, applfilters.FilterCommand(cmd))...,
		),
	)
}

func (d *Dispatcher) Handlers() []handlers.EventHandler {
	return d.handlers
}
