package ctx

import (
	"context"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/event"
)

type Ctx interface {
	Event() *event.Event
	Context() context.Context

	Get(key string) (any, error)

	Thread() *threads.MessagesThread
	SetThread(*threads.MessagesThread)

	Bot() bot.BotMessaging

	Answer(messages.Message) error

	IsHandled() bool
	SetHandled()
}
