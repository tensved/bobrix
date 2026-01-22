package commands

import (
	domcommands "github.com/tensved/bobrix/mxbot/domain/commands"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
)

// NewCommand - Command constructor
// name - name of the command
// commandHandler - handler of the command
// config - config of the command (optional)
func NewCommand(
	name string,
	handler func(domcommands.CommandCtx) error,
	config ...domcommands.CommandConfig,
) *domcommands.Command {
	cfg := domcommands.CommandConfig{
		Prefix:      domcommands.DefaultCommandPrefix,
		Description: map[string]string{},
	}
	if len(config) > 0 {
		cfg = config[0]
	}

	return &domcommands.Command{
		Prefix:      cfg.Prefix,
		Name:        name,
		Description: cfg.Description,
		Handler: func(c domctx.Ctx) error {
			commandCtx := NewCommandCtx(c)
			return handler(commandCtx)
		},
	}
}
