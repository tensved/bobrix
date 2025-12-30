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
	GetString(key string) (string, error)
	GetInt(key string) (int, error)

	Thread() *threads.MessagesThread
	SetThread(thread *threads.MessagesThread)

	Bot() bot.BotMessaging

	Answer(msg messages.Message) error
	TextAnswer(text string) error
	ErrorAnswer(errorText string, errorType int) error

	IsHandled() bool
	SetHandled()
	IsHandledWithUnlocker() (bool, func())
}
