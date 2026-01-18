package ctx

import (
	"context"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	domainctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"maunium.net/go/mautrix/event"
)

// Ensure Factory implements domain CtxFactory
var _ domainctx.CtxFactory = (*defaultFactory)(nil)

type defaultFactory struct {
	bot     domainbot.BotMessaging
	threads domainbot.BotThreads
	events  domainbot.EventLoader
}

func NewFactory(
	bot domainbot.BotMessaging,
	threads domainbot.BotThreads,
	events domainbot.EventLoader,
) domainctx.CtxFactory {
	return &defaultFactory{
		bot:     bot,
		threads: threads,
		events:  events,
	}
}

func (f *defaultFactory) New(
	ctx context.Context,
	evt *event.Event,
) (domainctx.Ctx, error) {
	return NewDefaultCtx(ctx, evt, f.bot, f.threads, f.events)
}
