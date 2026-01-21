package ctx

import (
	"context"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/botctx"
	domainctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"maunium.net/go/mautrix/event"
)

var _ domainctx.CtxFactory = (*defaultFactory)(nil)

type defaultFactory struct {
	bot     domainbot.BotMessaging
	threads domainbot.BotThreads
	events  domainbot.EventLoader
	// info    domainbot.BotInfo
	botCtx  botctx.Bot
}

func NewFactory(
	bot domainbot.BotMessaging,
	threads domainbot.BotThreads,
	events domainbot.EventLoader,
	// info domainbot.BotInfo,
	botCtx botctx.Bot,
) domainctx.CtxFactory {
	return &defaultFactory{
		bot:     bot,
		threads: threads,
		events:  events,
		// info:    info,
		botCtx: botCtx,
	}
}

func (f *defaultFactory) New(
	ctx context.Context,
	evt *event.Event,
) (domainctx.Ctx, error) {
	return NewDefaultCtx(ctx, evt, f.bot, f.botCtx, f.threads, f.events)
}
