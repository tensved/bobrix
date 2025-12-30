package ctx // ok

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tensved/bobrix/mxbot/domain/ctx"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/event"
	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ ctx.Ctx = (*DefaultCtx)(nil)

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
	context context.Context

	bot domainbot.BotMessaging

	thread *threads.MessagesThread

	storage map[string]any
	mx      *sync.Mutex

	handlesStatus *handlesStatus
}

type MetadataKey struct{}

var MetadataKeyContext = MetadataKey{}

func injectMetadataInContext(
	ctx context.Context,
	evt *event.Event,
	loader domainbot.EventLoader,
) context.Context {
	metadata := map[string]any{
		"event": evt,
	}

	if evt == nil {
		slog.Warn("event is nil, skipping metadata injection")
		return context.WithValue(ctx, MetadataKeyContext, metadata)
	}

	if msg := evt.Content.AsMessage(); msg != nil {
		if rel := msg.RelatesTo; rel != nil {
			metadata["thread_id"] = rel.EventID

			if loader != nil {
				mainEvent, err := loader.GetEvent(ctx, evt.RoomID, rel.EventID)
				if err != nil {
					slog.Error("error get main event", "error", err)
				}

				if mainEvent != nil {
					if answerTo, ok := mainEvent.Content.Raw[AnswerToCustomField]; ok {
						metadata["thread.answer_to"] = answerTo
					}
				} else {
					slog.Warn("main event is nil, skipping thread.answer_to")
				}
			}
		}
	}

	return context.WithValue(ctx, MetadataKeyContext, metadata)
}

func NewDefaultCtx(
	ctx context.Context,
	event *event.Event,
	bot domainbot.BotMessaging,
	threadProvider domainbot.Threads,
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
		context: ctx,
		event:   event,
		bot:     bot,
		thread:  thread,
		storage: map[string]any{},
		mx:      &sync.Mutex{},
		handlesStatus: &handlesStatus{
			mx: &sync.RWMutex{},
		},
	}, nil
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
	return c.storage[key].(int), nil
}

// Bot - get the bot from the context
func (c *DefaultCtx) Bot() domainbot.BotMessaging {
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
