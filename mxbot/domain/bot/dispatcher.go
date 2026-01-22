package bot

import (
	dfilters "github.com/tensved/bobrix/mxbot/domain/filters"
	dhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"
)

type EventDispatcher interface {
	AddEventHandler(handler dhandlers.EventHandler)
	AddFilter(f dfilters.Filter)
}
