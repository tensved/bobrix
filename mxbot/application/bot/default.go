package bot

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	applfilters "github.com/tensved/bobrix/mxbot/application/filters"
	"github.com/tensved/bobrix/mxbot/domain/bot"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/constructor"
	"maunium.net/go/mautrix/event"
)

var _ bot.FullBot = (*DefaultBot)(nil)

// Application-level coordinator
type DefaultBot struct {
	// --- identity
	name        string
	displayName string

	credentials     *config.BotCredentials
	botStatus       event.Presence
	isThreadEnabled bool

	// --- application layer
	dispatcher bot.EventDispatcher
	ctxFactory domctx.CtxFactory

	// --- infrastructure facades
	info        bot.BotInfo
	messaging   bot.BotMessaging
	threads     bot.BotThreads
	crypto      bot.BotCrypto
	client      bot.BotClient
	eventLoader bot.EventLoader
	rooms       bot.BotRoomActions
	typing      bot.BotTyping
	sync        bot.BotSync
	auth        bot.BotAuth
	health      bot.BotHealth
	media       bot.BotMedia
	persence    bot.BotPresenceControl

	// --- runtime state
	logger *zerolog.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// --- options
	syncerTimeRetry time.Duration
	typingTimeout   time.Duration

	// --- filters
	filters []domfilters.Filter

	// authMode config.AuthMode
	// asToken  string
}

// NewDefaultBot - Bot constructor
// botName - name of the bot (should be unique for engine)
// botCredentials - matrix credentials of the bot
func NewDefaultBot(
	displayName string,
	facade *constructor.MatrixBot, // infrastructure facade
	// dispatcher bot.EventDispatcher,
	// ctxFactory domctx.CtxFactory,
	logger *zerolog.Logger,
	credentials *config.BotCredentials,
	// botStatus event.Presence,
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

		dispatcher: facade,
		ctxFactory: facade,

		logger: logger,
		ctx:    ctx,
		cancel: cancel,

		syncerTimeRetry: 5 * time.Second,
		typingTimeout:   3 * time.Second,
	}

	defaultFilters := []domfilters.Filter{
		applfilters.FilterNotMe(facade), // ignore messages from the bot itself
		applfilters.FilterAfterStart(
			b.info,
			b.rooms,
			applfilters.FilterAfterStartOptions{
				StartTime:      time.Now(),
				ProcessInvites: true,
			},
		),
	}

	b.filters = defaultFilters

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *DefaultBot) Info() bot.BotInfo                { return b.info }
func (b *DefaultBot) Messaging() bot.BotMessaging      { return b.messaging }
func (b *DefaultBot) Threads() bot.BotThreads          { return b.threads }
func (b *DefaultBot) Crypto() bot.BotCrypto            { return b.crypto }
func (b *DefaultBot) Client() bot.BotClient            { return b.client }
func (b *DefaultBot) EventLoader() bot.EventLoader     { return b.eventLoader }
func (b *DefaultBot) Rooms() bot.BotRoomActions        { return b.rooms }
func (b *DefaultBot) Typing() bot.BotTyping            { return b.typing }
func (b *DefaultBot) Sync() bot.BotSync                { return b.sync }
func (b *DefaultBot) Auth() bot.BotAuth                { return b.auth }
func (b *DefaultBot) Health() bot.BotHealth            { return b.health }
func (b *DefaultBot) Media() bot.BotMedia              { return b.media }
func (b *DefaultBot) Presence() bot.BotPresenceControl { return b.persence }

