package mxbot

import (
	"context"
	"log/slog"
	"sync"

	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/event"
)

const (
	BobrixCustomField   = "bobrix"
	AnswerToCustomField = BobrixCustomField + ".answer_to"
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
	SetThread(thread *MessagesThread)

	Bot() Bot

	Answer(msg messages.Message) error
	TextAnswer(text string) error

	IsHandled() bool
	SetHandled()
	IsHandledWithUnlocker() (bool, func())
}

var _ Ctx = (*DefaultCtx)(nil)

type handlesStatus struct {
	isHandled bool
	mx        *sync.RWMutex
}

func (s *handlesStatus) IsHandledWithUnlocker() (bool, func()) {
	s.mx.RLock()

	if s.isHandled {
		s.mx.RUnlock()
		return true, func() {}
	}

	return false, func() {
		s.isHandled = true
		s.mx.RUnlock()
	}
}

func (c *DefaultCtx) IsHandledWithUnlocker() (bool, func()) {
	return c.handlesStatus.IsHandledWithUnlocker()
}

func (s *handlesStatus) Check() bool {
	return s.isHandled
}

func (s *handlesStatus) Set(status bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.isHandled = status
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

type MetadataKey struct{}

var MetadataKeyContext = MetadataKey{}

func injectMetadataInContext(ctx context.Context, evt *event.Event, bot Bot) context.Context {
	metadata := map[string]any{
		"event": evt,
	}

	if msg := evt.Content.AsMessage(); msg != nil {
		if rel := msg.RelatesTo; rel != nil {
			metadata["tread_id"] = rel.EventID

			mainEvent, err := bot.Client().GetEvent(ctx, evt.RoomID, rel.EventID)
			if err != nil {
				slog.Error("error get main event", "error", err)
			}

			metadata["thread.answer_to"] = mainEvent.Content.Raw[AnswerToCustomField]
		}
	}

	return context.WithValue(ctx, MetadataKeyContext, metadata)
}

func NewDefaultCtx(ctx context.Context, event *event.Event, bot Bot) (*DefaultCtx, error) {

	var thread *MessagesThread

	if bot.IsThreadEnabled() {
		var err error
		thread, err = bot.GetThreadByEvent(ctx, event)
		if err != nil {
			return nil, err
		}
	}

	ctx = injectMetadataInContext(ctx, event, bot)

	defaultCtx := &DefaultCtx{
		context: ctx,
		event:   event,
		storage: make(map[string]any),
		thread:  thread,
		bot:     bot,
		mx:      &sync.Mutex{},
		handlesStatus: &handlesStatus{
			isHandled: false,
			mx:        &sync.RWMutex{},
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

	msg.AddCustomFields(AnswerToCustomField, c.event.ID)

	return c.bot.SendMessage(c.Context(), c.event.RoomID, msg)
}

// TextAnswer - send a text message to the room
func (c *DefaultCtx) TextAnswer(text string) error {
	return c.Answer(messages.NewText(text))
}

// ErrorAnswer - send a text error message with error type added to the room
func (c *DefaultCtx) ErrorAnswer(errorText string, errorType string) error {
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

func (c *DefaultCtx) Thread() *MessagesThread {
	return c.thread
}

func (c *DefaultCtx) SetThread(thread *MessagesThread) {
	c.thread = thread
}
