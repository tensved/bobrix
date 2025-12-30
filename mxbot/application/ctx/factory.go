package ctx

import (
	"context"
	"sync"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"maunium.net/go/mautrix/event"
)

// type Factory interface {
// 	New(ctx context.Context, evt *event.Event) (dctx.Ctx, error)
// }

func NewDefaultCtx(
	ctx context.Context,
	event *event.Event,
	bot domainbot.BotMessaging,
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
		context: ctx,
		event:   event,
		bot:     bot,
		thread:  thread,
		storage: map[string]any{},
		mx:      &sync.Mutex{},
		handlesStatus: &handlesStatus{
			isHandled: false,
			mx:        sync.Mutex{},
		},
	}, nil
}
