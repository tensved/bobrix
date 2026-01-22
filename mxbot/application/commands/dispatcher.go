package commands

import (
	applfilters "github.com/tensved/bobrix/mxbot/application/filters"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domcommands "github.com/tensved/bobrix/mxbot/domain/commands"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"
	domhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"
)

type Dispatcher struct {
	handlers []domhandlers.EventHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: []domhandlers.EventHandler{},
	}
}

func Register(
	dispatcher dombot.EventDispatcher,
	cmd *domcommands.Command,
	extraFilters ...domfilters.Filter,
) {
	dispatcher.AddEventHandler(
		domhandlers.NewMessageHandler(
			cmd.Handler,
			append(extraFilters, applfilters.FilterCommand(cmd))...,
		),
	)
}

func (d *Dispatcher) Handlers() []domhandlers.EventHandler {
	return d.handlers
}
