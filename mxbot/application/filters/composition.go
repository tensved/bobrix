package filters

import (
	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/filters"
)

// FilterAny - filter for any of the given filters
// return true if any of the filters return true
func FilterAny(fs ...filters.Filter) filters.Filter {
	return func(evt *event.Event) bool {
		for _, f := range fs {
			if f(evt) {
				return true
			}
		}
		return false
	}
}

// FilterAll - filter for all of the given filters
// return true if all of the filters return true
func FilterAll(fs ...filters.Filter) filters.Filter {
	return func(evt *event.Event) bool {
		for _, f := range fs {
			if !f(evt) {
				return false
			}
		}
		return true
	}
}
