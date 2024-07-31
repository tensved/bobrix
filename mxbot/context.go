package mxbot

import (
	"context"
	"maunium.net/go/mautrix/event"
	"sync"
)

// Ctx - context of the bot
// provides access to the bot and event
// and allows to set and get values in the context
// and to answer to the event
type Ctx interface {
	Event() *event.Event

	Get(key string) (any, error)
	GetString(key string) (string, error)
	GetInt(key string) (int, error)

	Context() context.Context

	Bot() Bot

	Answer(text string) error
}

var _ Ctx = (*DefaultCtx)(nil)

type DefaultCtx struct {
	event  *event.Event
	values map[string]any
	mx     *sync.Mutex

	context context.Context

	bot Bot
}

func NewDefaultCtx(ctx context.Context, event *event.Event, bot Bot) *DefaultCtx {
	return &DefaultCtx{
		context: ctx,
		event:   event,
		values:  make(map[string]any),
		bot:     bot,
		mx:      &sync.Mutex{},
	}
}

func (c *DefaultCtx) Event() *event.Event {
	return c.event
}

func (c *DefaultCtx) Get(key string) (any, error) {
	return c.values[key], nil
}

func (c *DefaultCtx) GetString(key string) (string, error) {
	return c.values[key].(string), nil
}

func (c *DefaultCtx) GetInt(key string) (int, error) {
	return c.values[key].(int), nil
}

func (c *DefaultCtx) Bot() Bot {
	return c.bot
}

func (c *DefaultCtx) Answer(text string) error {

	return c.bot.SendMessage(c.Context(), &Message{
		RoomID: c.event.RoomID,
		Text:   text,
	})
}

func (c *DefaultCtx) Context() context.Context {
	return c.context
}
