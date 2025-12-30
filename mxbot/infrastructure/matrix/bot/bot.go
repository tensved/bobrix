package bot  // ok?

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	domain "github.com/tensved/bobrix/mxbot/domain/bot"

	// "github.com/tensved/bobrix/mxbot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
)

// service aggregator
// type DefaultBot struct {
// 	name   string
// 	client *mautrix.Client
// 	logger *zerolog.Logger

// 	auth    *authService
// 	crypto  *cryptoService
// 	syncer  *syncService
// 	typing  *typingService
// 	threads *threadService
// 	history *historyService
// 	loader  *EventLoaderService

// 	filters  []mxbot.Filter
// 	handlers []mxbot.EventHandler

// 	cancel context.CancelFunc
// }

type DefaultBot struct {
	name        string
	displayName string

	matrixClient  *mautrix.Client
	machine *crypto.OlmMachine

	logger      *zerolog.Logger
	credentials *BotCredentials

	syncerRetry   time.Duration
	typingTimeout time.Duration

	cancel context.CancelFunc

	isThreadEnabled bool
}

// BotCredentials - credentials of the bot for Matrix
// should be provided by the user
// (username, password, homeserverURL)
type BotCredentials struct {
	Username      string
	Password      string
	HomeServerURL string
	PickleKey     []byte
	ThreadLimit   int
	AuthMode      AuthMode
	ASToken       string
}

var (
	defaultSyncerRetryTime = 5 * time.Second
	defaultTypingTimeout   = 30 * time.Second
)

type BotOptions func(*DefaultBot) // Bot options. Used to configure the bot

// var _ mxbot.Bot = (*DefaultBot)(nil)

var _ domain.BotInfo = (*DefaultBot)(nil)
var _ domain.BotClient = (*DefaultBot)(nil)
var _ domain.BotMessaging = (*DefaultBot)(nil)
var _ domain.BotThreads = (*DefaultBot)(nil)
var _ domain.EventLoader = (*DefaultBot)(nil)
