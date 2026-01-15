package bot

import dhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"

type EventDispatcher interface {
	AddEventHandler(handler dhandlers.EventHandler)
}
