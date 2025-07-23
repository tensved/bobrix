package mxbot

import (
	"maunium.net/go/mautrix/event"
	"strings"
)

const (
	DefaultCommandPrefix = "/"
)

// CommandConfig - config of the command (should be provided by the user)
// (prefix, description)
type CommandConfig struct {
	Prefix      string
	Description map[string]string
}

// Command - command of the bot
// Describes the command and its handler
type Command struct {
	Prefix      string
	Name        string
	Description map[string]string
	Handler     func(ctx Ctx) error
}

// CommandCtx - context of the command
// provides access to the command and its arguments
// is a wrapper for the Ctx
type CommandCtx interface {
	Ctx             // provides access to the Ctx
	Args() []string // provides access to the command arguments
}

// NewCommand - Command constructor
// name - name of the command
// commandHandler - handler of the command
// config - config of the command (optional)
func NewCommand(name string, commandHandler func(ctx CommandCtx) error, config ...CommandConfig) *Command {
	cfg := CommandConfig{
		Prefix:      DefaultCommandPrefix,
		Description: map[string]string{},
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	handler := func(ctx Ctx) error {
		commandCtx := NewCommandCtx(ctx)
		return commandHandler(commandCtx)
	}

	return &Command{
		Prefix:      cfg.Prefix,
		Name:        name,
		Description: cfg.Description,
		Handler:     handler,
	}
}

// NewCommandCtx - CommandCtx constructor
// wrap the Ctx
func NewCommandCtx(ctx Ctx) CommandCtx {

	return &DefaultCommandCtx{
		Ctx:  ctx,
		args: parseArgsFromEvent(ctx.Event()),
	}
}

// DefaultCommandCtx - CommandCtx implementation
// provides access to the command and its arguments
type DefaultCommandCtx struct {
	Ctx
	args []string
}

func (c *DefaultCommandCtx) Args() []string {
	return c.args
}

// parseArgsFromEvent - parse the arguments from the event
// (ignores the command prefix and command name)
func parseArgsFromEvent(evt *event.Event) []string {
	if evt.Type != event.EventMessage {
		return nil
	}

	wordsInMessage := strings.Split(evt.Content.AsMessage().Body, " ")

	if len(wordsInMessage) < 2 {
		return []string{}
	}

	args := make([]string, len(wordsInMessage)-1)

	copy(args, wordsInMessage[1:])

	return args
}
