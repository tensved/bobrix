package ctx

import (
	"context"
	"fmt"
	"sync"

	"maunium.net/go/mautrix/event"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	dombotctx "github.com/tensved/bobrix/mxbot/domain/botctx"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"github.com/tensved/bobrix/mxbot/messages"
)

var _ domctx.Ctx = (*DefaultCtx)(nil)

type DefaultCtx struct {
	event   *event.Event
	context context.Context

	botMessaging dombot.BotMessaging
	botCtx       dombotctx.Bot

	thread *threads.MessagesThread

	storage map[string]any
	mx      *sync.Mutex

	handlesStatus *handlesStatus
	cancel        context.CancelFunc

	claimMu sync.Mutex
	claimed bool
}

func NewDefaultCtx(
	ctx context.Context,
	event *event.Event,
	botMessaging dombot.BotMessaging,
	botCtx dombotctx.Bot,
	threadProvider dombot.BotThreads,
	eventLoader dombot.EventLoader,
) (*DefaultCtx, error) {
	var thread *threads.MessagesThread

	if threadProvider != nil && threadProvider.IsThreadEnabled() {
		var err error
		thread, err = threadProvider.GetThreadByEvent(ctx, event)
		if err != nil {
			return nil, err
		}
	}

	evtCtx, cancel := context.WithCancel(ctx)
	evtCtx = injectMetadataInContext(evtCtx, event, eventLoader)

	return &DefaultCtx{
		context:       evtCtx,
		cancel:        cancel,
		event:         event,
		botMessaging:  botMessaging,
		botCtx:        botCtx,
		thread:        thread,
		storage:       map[string]any{},
		mx:            &sync.Mutex{},
		handlesStatus: newHandlesStatus(),
	}, nil
}

func (c *DefaultCtx) Handled() <-chan struct{} {
	return c.handlesStatus.doneCh()
}

func (c *DefaultCtx) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *DefaultCtx) IsHandledWithUnlocker() (bool, func()) {
	return c.handlesStatus.isHandledWithUnlocker()
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
func (c *DefaultCtx) BotMessaging() dombot.BotMessaging {
	return c.botMessaging
}

func (c *DefaultCtx) Bot() dombotctx.Bot {
	return c.botCtx
}

// Answer - send a message to the room.
// it is a wrapper for bot.SendMessage
// it returns an error if the message could not be sent
func (c *DefaultCtx) Answer(msg messages.Message) error {
	err := c.Send(msg)
	if err == nil {
		c.SetHandled()
	}
	return err
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

// Send - send a message to the room without marking it "handled."
func (c *DefaultCtx) Send(msg messages.Message) error {
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

	msg.AddCustomFields(domctx.AnswerToCustomField, c.event.ID)
	return c.botMessaging.SendMessage(c.Context(), c.event.RoomID, msg)
}

func (c *DefaultCtx) TextSend(text string) error {
	return c.Send(messages.NewText(text))
}

// Context - return the root context
func (c *DefaultCtx) Context() context.Context {
	return c.context
}

func (c *DefaultCtx) IsHandled() bool {
	return c.handlesStatus.check()
}

func (c *DefaultCtx) SetHandled() {
	c.handlesStatus.set(true)
}

func (c *DefaultCtx) Thread() *threads.MessagesThread {
	return c.thread
}

func (c *DefaultCtx) SetThread(thread *threads.MessagesThread) {
	c.thread = thread
}

// TryClaim atomically "claims" the current event context for handling.
//
// Purpose:
//   - Prevent multiple handlers (e.g. contract parser + fallback) from processing/sending replies
//     for the same incoming Matrix event.
//   - This is separate from Handled(): claim means "I will handle it", handled means
//     "a final reply was actually sent".
//
// Semantics:
//   - Returns true only for the first caller.
//   - All subsequent callers get false.
//   - The claim is not released (one-shot) because the context is tied to a single event.
func (c *DefaultCtx) TryClaim() bool {
	c.claimMu.Lock()
	defer c.claimMu.Unlock()
	if c.claimed {
		return false
	}
	c.claimed = true
	return true
}

// IsClaimed reports whether the event context has already been claimed by some handler.
//
// Note:
//   - This does NOT mean a reply was sent. For that, use IsHandled()/Handled().
//   - Intended for early checks/short-circuiting in handlers that should run only
//     if no one else has taken ownership of the event.
func (c *DefaultCtx) IsClaimed() bool {
	c.claimMu.Lock()
	defer c.claimMu.Unlock()
	return c.claimed
}
