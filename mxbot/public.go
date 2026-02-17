package mxbot

import (
	"fmt"

	"maunium.net/go/mautrix/event"

	applbot "github.com/tensved/bobrix/mxbot/application/bot"
	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"
	domhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"
	domthreads "github.com/tensved/bobrix/mxbot/domain/threads"

	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	infrabot "github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
	infrathreads "github.com/tensved/bobrix/mxbot/infrastructure/matrix/threads"
)

type Bot = dombot.FullBot
type Filter = domfilters.Filter
type EventHandler = domhandlers.EventHandler
type MessagesThread = domthreads.MessagesThread
type BotMedia = dombot.BotMedia
type BotInfo = dombot.BotInfo
type BotClient = dombot.BotClient
type Config = infrabot.Config
type BotCredentials = infracfg.BotCredentials
type Ctx = domctx.Ctx
type BotOptions = applbot.BotOptions

var MetadataKeyContext = domctx.MetadataKeyContext

const AnswerToCustomField = domctx.AnswerToCustomField

func NewMatrixBot(cfg Config, opts ...applbot.BotOptions) (*applbot.DefaultBot, error) {
	facade, err := infrabot.NewMatrixBot(cfg)
	if err != nil {
		return nil, fmt.Errorf("err create facade: %w", err)
	}

	return applbot.NewDefaultBot(
		cfg.Credentials.Username,
		facade,
		cfg.Logger,
		cfg.Credentials,
		opts...,
	), nil
}

func WithDisplayName(name string) BotOptions {
	return applbot.WithDisplayName(name)
}

func GetFixedMessage(evt *event.Event) (*event.MessageEventContent, bool) {
	return infrathreads.GetFixedMessage(evt)
}
