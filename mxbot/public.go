package mxbot

import (
	"fmt"
	"time"

	// applctx "github.com/tensved/bobrix/mxbot/application/ctx"
	// "github.com/tensved/bobrix/mxbot"
	applbot "github.com/tensved/bobrix/mxbot/application/bot"
	// applhandlers "github.com/tensved/bobrix/mxbot/application/handlers"
	"maunium.net/go/mautrix/event"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"github.com/tensved/bobrix/mxbot/domain/threads"
	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	infrabot "github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
	infrathreads "github.com/tensved/bobrix/mxbot/infrastructure/matrix/threads"
)

type Bot = dombot.FullBot

// type Bot =
type Filter = filters.Filter
type EventHandler = handlers.EventHandler
type MessagesThread = threads.MessagesThread
type BotMedia = dombot.BotMedia
type BotInfo = dombot.BotInfo
type BotClient = dombot.BotClient
type Config = infrabot.Config
type BotCredentials = infracfg.BotCredentials

var MetadataKeyContext = domctx.MetadataKeyContext

const AnswerToCustomField = domctx.AnswerToCustomField

type Ctx = domctx.Ctx

// type Ctx interface {
// 	domctx.Ctx
// 	Bot() dombot.FullBot
// }

type BotOptions = applbot.BotOptions

func NewMatrixBot(cfg Config, opts ...applbot.BotOptions) (*applbot.DefaultBot, error) {
	facade, err := infrabot.NewMatrixBot(cfg, []filters.Filter{})
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

func WithSyncerRetryTime(d time.Duration) BotOptions {
	return applbot.WithSyncerRetryTime(d)
}

func WithTypingTimeout(d time.Duration) BotOptions {
	return applbot.WithTypingTimeout(d)
}

func WithDisplayName(name string) BotOptions {
	return applbot.WithDisplayName(name)
}

func GetFixedMessage(evt *event.Event) (*event.MessageEventContent, bool) {
	return infrathreads.GetFixedMessage(evt)
}
