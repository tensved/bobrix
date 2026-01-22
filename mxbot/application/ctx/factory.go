package ctx

import (
	"context"

	"maunium.net/go/mautrix/event"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	dombotctx "github.com/tensved/bobrix/mxbot/domain/botctx"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
)

var _ domctx.CtxFactory = (*defaultFactory)(nil)

type defaultFactory struct {
	bot     dombot.BotMessaging
	threads dombot.BotThreads
	events  dombot.EventLoader
	botCtx  dombotctx.Bot
}

func NewFactory(
	bot dombot.BotMessaging,
	threads dombot.BotThreads,
	events dombot.EventLoader,
	botCtx dombotctx.Bot,
) domctx.CtxFactory {
	return &defaultFactory{
		bot:     bot,
		threads: threads,
		events:  events,
		botCtx:  botCtx,
	}
}

func (f *defaultFactory) New(
	ctx context.Context,
	evt *event.Event,
) (domctx.Ctx, error) {
	return NewDefaultCtx(ctx, evt, f.bot, f.botCtx, f.threads, f.events)
}
