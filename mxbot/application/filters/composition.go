package filters

import (
	df "github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix/event"
)

// FilterAny - filter for any of the given filters
// return true if any of the filters return true
func FilterAny(fs ...df.Filter) df.Filter {
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
func FilterAll(fs ...df.Filter) df.Filter {
	return func(evt *event.Event) bool {
		for _, f := range fs {
			if !f(evt) {
				return false
			}
		}
		return true
	}
}
