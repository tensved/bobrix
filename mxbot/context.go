package mxbot

import (
	"context"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/event"
	"sync"
)

// Ctx - context of the bot
// provides access to the bot and event
// and allows to set and get storage in the context
// and to answer to the event
type Ctx interface {
	Event() *event.Event

	Get(key string) (any, error)
	GetString(key string) (string, error)
	GetInt(key string) (int, error)

	Context() context.Context

	Bot() Bot

	Answer(msg messages.Message) error
	TextAnswer(text string) error

	IsHandled() bool
	SetHandled()
	CheckAndSetHandled() bool
}

var _ Ctx = (*DefaultCtx)(nil)

type DefaultCtx struct {
	event   *event.Event
	storage map[string]any
	mx      *sync.Mutex

	context context.Context

	bot Bot

	isHandled bool
	handledMx *sync.Mutex
}

func NewDefaultCtx(ctx context.Context, event *event.Event, bot Bot) *DefaultCtx {
	return &DefaultCtx{
		context:   ctx,
		event:     event,
		storage:   make(map[string]any),
		bot:       bot,
		mx:        &sync.Mutex{},
		isHandled: false,
		handledMx: &sync.Mutex{},
	}
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

// GetString - get a string from the context
func (c *DefaultCtx) GetString(key string) (string, error) {
	return c.storage[key].(string), nil
}

// GetInt - get an int from the context
func (c *DefaultCtx) GetInt(key string) (int, error) {
	return c.storage[key].(int), nil
}

// Bot - get the bot from the context
func (c *DefaultCtx) Bot() Bot {
	return c.bot
}

// Answer - send a message to the room.
// it is a wrapper for bot.SendMessage
// it returns an error if the message could not be sent
func (c *DefaultCtx) Answer(msg messages.Message) error {
	return c.bot.SendMessage(c.Context(), c.event.RoomID, msg)
}

// TextAnswer - send a text message to the room
func (c *DefaultCtx) TextAnswer(text string) error {
	return c.Answer(messages.NewText(text))
}

// Context - return the root context
func (c *DefaultCtx) Context() context.Context {
	return c.context
}

func (c *DefaultCtx) IsHandled() bool {
	return c.isHandled
}

func (c *DefaultCtx) SetHandled() {
	c.isHandled = true
}

func (c *DefaultCtx) CheckAndSetHandled() bool {
	c.handledMx.Lock()
	defer c.handledMx.Unlock()

	if c.IsHandled() {
		return false
	}

	c.SetHandled()

	return true
}
