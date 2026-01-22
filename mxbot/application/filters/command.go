package filters

import (
	"strings"

	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/commands"
	"github.com/tensved/bobrix/mxbot/domain/filters"
)

// FilterCommand - filter for command messages
// (check if message starts with command prefix and name)
func FilterCommand(cmd *commands.Command) filters.Filter {
	return func(evt *event.Event) bool {
		if evt.Type != event.EventMessage {
			return false
		}

		words := strings.Split(evt.Content.AsMessage().Body, " ")
		if len(words) == 0 {
			return false
		}

		return strings.EqualFold(
			words[0],
			cmd.Prefix+cmd.Name,
		)
	}
}
