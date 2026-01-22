package commands

import (
	"github.com/tensved/bobrix/mxbot/domain/ctx"
)

// CommandCtx - context of the command
// provides access to the command and its arguments
// is a wrapper for the Ctx
type CommandCtx interface {
	ctx.Ctx         // provides access to the Ctx
	Args() []string // provides access to the command arguments
}
