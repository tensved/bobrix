package commands

import (
	"maunium.net/go/mautrix/event"

	applfilters "github.com/tensved/bobrix/mxbot/application/filters"

	domcommands "github.com/tensved/bobrix/mxbot/domain/commands"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"
	domhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"
)

func NewCommandEventHandler(
	cmd *domcommands.Command,
	extraFilters ...domfilters.Filter,
) domhandlers.EventHandler {
	allFilters := []domfilters.Filter{
		applfilters.FilterCommand(cmd),
	}
	allFilters = append(allFilters, extraFilters...)

	return domhandlers.NewEventHandler(
		event.EventMessage,
		func(c domctx.Ctx) error {
			return cmd.Handler(c)
		},
		allFilters...,
	)
}
