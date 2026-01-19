package mxbot

import (
	"errors"

	// applctx "github.com/tensved/bobrix/mxbot/application/ctx"
	// applhandlers "github.com/tensved/bobrix/mxbot/application/handlers"
	applbot "github.com/tensved/bobrix/mxbot/application/bot"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"

	// "github.com/tensved/bobrix/mxbot/domain/ctx"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	// "github.com/tensved/bobrix/mxbot/domain/filters"
	// "github.com/tensved/bobrix/mxbot/domain/handlers"
	"github.com/tensved/bobrix/mxbot/domain/threads"
	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	infrabot "github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
)

type Bot = dombot.FullBot
// type Bot = *applbot.DefaultBot

type Filter = filters.Filter
type EventHandler = handlers.EventHandler
type MessagesThread = threads.MessagesThread
type BotMedia = dombot.BotMedia
type Config = infrabot.Config
type BotCredentials = infracfg.BotCredentials

type Ctx interface {
	domctx.Ctx
	Bot() dombot.BotInfo
}

func NewMatrixBot(cfg Config) (*applbot.DefaultBot, error) {
	facade, err := infrabot.NewMatrixBot(cfg, []filters.Filter{})
	if err != nil {
		return nil, errors.New("err create facade")
	}

	return applbot.NewDefaultBot(
		cfg.Credentials.Username,
		facade,
		cfg.Logger,
		cfg.Credentials,
	), nil
}
