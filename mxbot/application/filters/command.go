package filters

import (
	"strings"

	df "github.com/tensved/bobrix/mxbot/domain/filters"
	dc "github.com/tensved/bobrix/mxbot/domain/commands"
	"maunium.net/go/mautrix/event"
)

// FilterCommand - filter for command messages
// (check if message starts with command prefix and name)
func FilterCommand(cmd *dc.Command) df.Filter {
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
