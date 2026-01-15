package bot

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
)

type DefaultBot struct {
	name        string
	displayName string

	matrixClient *mautrix.Client
	machine      *crypto.OlmMachine

	logger      *zerolog.Logger
	credentials *config.BotCredentials

	syncerRetry   time.Duration
	typingTimeout time.Duration

	cancel context.CancelFunc

	isThreadEnabled bool

	botStatus event.Presence

	authMode config.AuthMode
	asToken  string
}

// type DefaultBot2 struct {
// 	eventHandlers []EventHandler

// 	filters []Filter
// }

var (
	defaultSyncerRetryTime = 5 * time.Second
	defaultTypingTimeout   = 30 * time.Second
)

// type BotOptions func(*DefaultBot) // Bot options. Used to configure the bot

// func NewDefault(
// 	name string,
// 	creds *BotCredentials,
// 	opts ...BotOptions,
// ) (*DefaultBot, error) {
// 	return &DefaultBot{}, nil
// }
