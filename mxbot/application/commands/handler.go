package commands

import (
	f "github.com/tensved/bobrix/mxbot/application/filters"
	dc "github.com/tensved/bobrix/mxbot/domain/commands"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	df "github.com/tensved/bobrix/mxbot/domain/filters"
	dh "github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

func NewCommandEventHandler(
	cmd *dc.Command,
	extraFilters ...df.Filter,
) dh.EventHandler {
	allFilters := []df.Filter{
		f.FilterCommand(cmd),
	}
	allFilters = append(allFilters, extraFilters...)

	return dh.NewEventHandler(
		event.EventMessage,
		func(c ctx.Ctx) error {
			return cmd.Handler(c)
		},
		allFilters...,
	)
}
