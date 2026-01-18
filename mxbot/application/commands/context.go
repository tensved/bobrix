package commands

import (
	"strings"

	"github.com/tensved/bobrix/mxbot/domain/commands"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"maunium.net/go/mautrix/event"
)

// DefaultCommandCtx - CommandCtx implementation
// provides access to the command and its arguments
type defaultCommandCtx struct {
	ctx.Ctx
	args []string
}

func (c *defaultCommandCtx) Args() []string {
	return c.args
}

// NewCommandCtx - CommandCtx constructor
// wrap the Ctx
func NewCommandCtx(c ctx.Ctx) commands.CommandCtx {
	return &defaultCommandCtx{
		Ctx:  c,
		args: parseArgsFromEvent(c.Event()),
	}
}

// parseArgsFromEvent - parse the arguments from the event
// (ignores the command prefix and command name)
func parseArgsFromEvent(evt *event.Event) []string {
	if evt.Type != event.EventMessage {
		return nil
	}

	words := strings.Split(evt.Content.AsMessage().Body, " ")
	if len(words) < 2 {
		return []string{}
	}

	return words[1:]
}
