package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/event"

	appldispatcher "github.com/tensved/bobrix/mxbot/application/dispatcher"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	dombotctx "github.com/tensved/bobrix/mxbot/domain/botctx"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"

	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	infraconstr "github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
)

var _ dombot.FullBot = (*DefaultBot)(nil)
var _ dombotctx.Bot = (*DefaultBot)(nil)

// Application-level coordinator
type DefaultBot struct {
	// --- identity
	name        string
	displayName string

	credentials     *infracfg.BotCredentials
	botStatus       event.Presence
	isThreadEnabled bool

	// --- application layer
	dispatcher *appldispatcher.Dispatcher
	ctxFactory domctx.CtxFactory

	// --- infrastructure facades
	info        dombot.BotInfo
	messaging   dombot.BotMessaging
	threads     dombot.BotThreads
	crypto      dombot.BotCrypto
	client      dombot.BotClient
	eventLoader dombot.EventLoader
	rooms       dombot.BotRoomActions
	typing      dombot.BotTyping
	sync        dombot.BotSync
	auth        dombot.BotAuth
	health      dombot.BotHealth
	media       dombot.BotMedia
	persence    dombot.BotPresenceControl

	// --- runtime state
	logger *zerolog.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// --- options
	syncerTimeRetry time.Duration
	typingTimeout   time.Duration

	// --- filters
	filters []domfilters.Filter
}

// NewDefaultBot - Bot constructor
// botName - name of the bot (should be unique for engine)
// botCredentials - matrix credentials of the bot
func NewDefaultBot(
	displayName string,
	facade *infraconstr.MatrixBot, // infra facade
	logger *zerolog.Logger,
	credentials *infracfg.BotCredentials,
	opts ...BotOptions,
) *DefaultBot {
	ctx, cancel := context.WithCancel(context.Background())

	b := &DefaultBot{
		name:        credentials.Username,
		displayName: displayName,

		credentials:     credentials,
		botStatus:       event.PresenceOnline,
		isThreadEnabled: credentials.IsThreadEnabled,

		info:        facade,
		messaging:   facade,
		threads:     facade,
		crypto:      facade,
		client:      facade,
		eventLoader: facade,
		rooms:       facade,
		typing:      facade,
		sync:        facade,
		auth:        facade,
		health:      facade,
		media:       facade,
		persence:    facade,

		dispatcher: facade.Dispatcher,
		ctxFactory: facade.CtxFactory,

		logger: logger,
		ctx:    ctx,
		cancel: cancel,

		syncerTimeRetry: 5 * time.Second,
		typingTimeout:   3 * time.Second,
	}

	b.filters = b.dispatcher.Filters()

	for _, opt := range opts {
		opt(b)
	}

	slog.Info("New bot created", "bot", b.name)

	return b
}

func (b *DefaultBot) Info() dombot.BotInfo                { return b.info }
func (b *DefaultBot) Messaging() dombot.BotMessaging      { return b.messaging }
func (b *DefaultBot) Threads() dombot.BotThreads          { return b.threads }
func (b *DefaultBot) Crypto() dombot.BotCrypto            { return b.crypto }
func (b *DefaultBot) Client() dombot.BotClient            { return b.client }
func (b *DefaultBot) EventLoader() dombot.EventLoader     { return b.eventLoader }
func (b *DefaultBot) Rooms() dombot.BotRoomActions        { return b.rooms }
func (b *DefaultBot) Typing() dombot.BotTyping            { return b.typing }
func (b *DefaultBot) Sync() dombot.BotSync                { return b.sync }
func (b *DefaultBot) Auth() dombot.BotAuth                { return b.auth }
func (b *DefaultBot) Health() dombot.BotHealth            { return b.health }
func (b *DefaultBot) Media() dombot.BotMedia              { return b.media }
func (b *DefaultBot) Presence() dombot.BotPresenceControl { return b.persence }
