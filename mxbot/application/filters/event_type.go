package filters

import (
	"slices"

	df "github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix/event"
)

// FilterEventTypes - filter for specific event types
// check if event type is in the list
// return true if event type is in the list
func FilterEventTypes(types ...event.Type) df.Filter {
	return func(evt *event.Event) bool {
		return slices.Contains(types, evt.Type)
	}
}

// FilterEventMessage - filter for event messages
// check if event type is event.EventMessage
// return true if event type is event.EventMessage
func FilterEventMessage() df.Filter {
	return FilterEventTypes(event.EventMessage)
}
