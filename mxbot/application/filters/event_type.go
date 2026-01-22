package filters

import (
	"slices"

	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/filters"
)

// FilterEventTypes - filter for specific event types
// check if event type is in the list
// return true if event type is in the list
func FilterEventTypes(types ...event.Type) filters.Filter {
	return func(evt *event.Event) bool {
		return slices.Contains(types, evt.Type)
	}
}

// FilterEventMessage - filter for event messages
// check if event type is event.EventMessage
// return true if event type is event.EventMessage
func FilterEventMessage() filters.Filter {
	return FilterEventTypes(event.EventMessage)
}
