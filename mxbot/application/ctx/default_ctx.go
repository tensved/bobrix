package ctx

import (
	"context"
	"fmt"
	"sync"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/botctx"
	domainctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/event"
	// "maunium.net/go/mautrix/id"
)

var _ domainctx.Ctx = (*DefaultCtx)(nil)

// var _ botctx.Bot = (*DefaultCtx)(nil)

type DefaultCtx struct {
	event   *event.Event
	context context.Context

	botMessaging domainbot.BotMessaging
	botCtx       botctx.Bot

	thread *threads.MessagesThread

	storage map[string]any
	mx      *sync.Mutex

	handlesStatus *handlesStatus
}

func NewDefaultCtx(
	ctx context.Context,
	event *event.Event,
	botMessaging domainbot.BotMessaging,
	botCtx botctx.Bot,
	threadProvider domainbot.BotThreads,
	eventLoader domainbot.EventLoader,
) (*DefaultCtx, error) {
	var thread *threads.MessagesThread

	if threadProvider != nil && threadProvider.IsThreadEnabled() {
		var err error
		thread, err = threadProvider.GetThreadByEvent(ctx, event)
		if err != nil {
			return nil, err
		}
	}

	ctx = injectMetadataInContext(ctx, event, eventLoader)

	return &DefaultCtx{
		context:      ctx,
		event:        event,
		botMessaging: botMessaging,
		botCtx:       botCtx,
		thread:       thread,
		storage:      map[string]any{},
		mx:           &sync.Mutex{},
		handlesStatus: &handlesStatus{
			isHandled: false,
			mx:        sync.Mutex{},
		},
	}, nil
}

func (c *DefaultCtx) IsHandledWithUnlocker() (bool, func()) {
	return c.handlesStatus.IsHandledWithUnlocker()
}

// Event - get the event from the context.
// it is a wrapper for the event
func (c *DefaultCtx) Event() *event.Event {
	return c.event
}

// Get - get a value from the context
func (c *DefaultCtx) Get(key string) (any, error) {
	return c.storage[key], nil
}

// GetString - get a string from the context safely
func (c *DefaultCtx) GetString(key string) (string, error) {
	val, ok := c.storage[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("value for key %q is not a string", key)
	}
	return str, nil
}

// GetInt - get an int from the context
func (c *DefaultCtx) GetInt(key string) (int, error) {
	v, ok := c.storage[key]
	if !ok {
		return 0, fmt.Errorf("key %q not found", key)
	}
	i, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("value for key %q is not int", key)
	}
	return i, nil
}

// Bot - get the bot from the context
func (c *DefaultCtx) BotMessaging() domainbot.BotMessaging {
	return c.botMessaging
}

func (c *DefaultCtx) Bot() botctx.Bot {
	return c.botCtx
}

// Answer - send a message to the room.
// it is a wrapper for bot.SendMessage
// it returns an error if the message could not be sent
func (c *DefaultCtx) Answer(msg messages.Message) error {
	thread := c.thread
	if thread != nil {
		msg.SetRelatesTo(&event.RelatesTo{
			Type:    event.RelThread,
			EventID: thread.ParentID,
			InReplyTo: &event.InReplyTo{
				EventID: c.Event().ID,
			},
			IsFallingBack: true,
		})
	}

	msg.AddCustomFields(domainctx.AnswerToCustomField, c.event.ID)

	return c.botMessaging.SendMessage(c.Context(), c.event.RoomID, msg)
}

// TextAnswer - send a text message to the room
func (c *DefaultCtx) TextAnswer(text string) error {
	return c.Answer(messages.NewText(text))
}

// ErrorAnswer - send a text error message to the room with error_code added
func (c *DefaultCtx) ErrorAnswer(errorText string, errorType int) error {
	msg := messages.NewText(errorText)
	msg.AddCustomFields("error_code", errorType)
	return c.Answer(msg)
}

// Context - return the root context
func (c *DefaultCtx) Context() context.Context {
	return c.context
}

func (c *DefaultCtx) IsHandled() bool {
	return c.handlesStatus.Check()
}

func (c *DefaultCtx) SetHandled() {
	c.handlesStatus.Set(true)
}

func (c *DefaultCtx) Thread() *threads.MessagesThread {
	return c.thread
}

func (c *DefaultCtx) SetThread(thread *threads.MessagesThread) {
	c.thread = thread
}
