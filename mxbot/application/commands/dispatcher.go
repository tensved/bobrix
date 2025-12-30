package commands

import (
	dc "github.com/tensved/bobrix/mxbot/domain/commands"
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

func (d *Dispatcher) Register(
	cmd *dc.Command,
	filters ...filters.Filter,
) {
	h := NewCommandEventHandler(cmd, filters...)
	d.handlers = append(d.handlers, h)
}

func (d *Dispatcher) Handlers() []handlers.EventHandler {
	return d.handlers
}
