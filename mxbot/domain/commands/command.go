package commands //!ok!

import "github.com/tensved/bobrix/mxbot/domain/ctx"

const DefaultCommandPrefix = "/"

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
	Handler     func(ctx.Ctx) error
}
