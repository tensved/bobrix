package ctx

import (
	domain "github.com/tensved/bobrix/mxbot/domain/bot"
)

type BotCtx struct {
	domain.BotClient
	domain.BotInfo
}

func NewBotCtx(c domain.BotClient, i domain.BotInfo) BotCtx {
	return BotCtx{
		BotClient: c,
		BotInfo:   i,
	}
}
