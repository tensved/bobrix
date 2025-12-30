package commands

import (
	dc "github.com/tensved/bobrix/mxbot/domain/commands"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
)

// NewCommand - Command constructor
// name - name of the command
// commandHandler - handler of the command
// config - config of the command (optional)
func NewCommand(
	name string,
	handler func(dc.CommandCtx) error,
	config ...dc.CommandConfig,
) *dc.Command {
	cfg := dc.CommandConfig{
		Prefix:      dc.DefaultCommandPrefix,
		Description: map[string]string{},
	}
	if len(config) > 0 {
		cfg = config[0]
	}

	return &dc.Command{
		Prefix:      cfg.Prefix,
		Name:        name,
		Description: cfg.Description,
		Handler: func(c ctx.Ctx) error {
			commandCtx := NewCommandCtx(c)
			return handler(commandCtx)
		},
	}
}
