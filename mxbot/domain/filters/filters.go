package filters

import (
	"maunium.net/go/mautrix/event"
)

// Filter - message filter
// return true if message should be processed
// return false if message should be ignored
type Filter func(evt *event.Event) bool
