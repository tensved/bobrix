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

	Thread() *MessagesThread

	Bot() Bot

	Answer(msg messages.Message) error
	TextAnswer(text string) error

	IsHandled() bool
	SetHandled()
	CheckAndSetHandled() bool
}

var _ Ctx = (*DefaultCtx)(nil)

type handlesStatus struct {
	isHandled bool
	mx        *sync.Mutex
}

func (s *handlesStatus) Check() bool {
	return s.isHandled
}

func (s *handlesStatus) Set(status bool) {
	s.isHandled = status
}

func (s *handlesStatus) Lock() {
	s.mx.Lock()
}

func (s *handlesStatus) Unlock() {
	s.mx.Unlock()
}

type DefaultCtx struct {
	event   *event.Event
	storage map[string]any
	thread  *MessagesThread
	mx      *sync.Mutex

	context context.Context

	bot Bot

	handlesStatus *handlesStatus
}

func NewDefaultCtx(ctx context.Context, event *event.Event, bot Bot) (*DefaultCtx, error) {

	thread, err := bot.GetThreadByEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	defaultCtx := &DefaultCtx{
		context: ctx,
		event:   event,
		storage: make(map[string]any),
		thread:  thread,
		bot:     bot,
		mx:      &sync.Mutex{},
		handlesStatus: &handlesStatus{
			isHandled: false,
			mx:        &sync.Mutex{},
		},
	}

	return defaultCtx, nil
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
	return c.handlesStatus.Check()
}

func (c *DefaultCtx) SetHandled() {
	c.handlesStatus.Set(true)
}

func (c *DefaultCtx) CheckAndSetHandled() bool {
	c.handlesStatus.Lock()
	defer c.handlesStatus.Unlock()

	if c.IsHandled() {
		return false
	}

	c.SetHandled()

	return true
}

func (c *DefaultCtx) Thread() *MessagesThread {
	return c.thread
}
