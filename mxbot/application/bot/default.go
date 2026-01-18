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
	info      bot.BotInfo
	messaging bot.BotMessaging
	threads   bot.BotThreads
	crypto    bot.BotCrypto
	rooms     bot.BotRoomActions
	typing    bot.BotTyping
	sync      bot.BotSync
	client    bot.BotClient
	auth      bot.BotAuth
	health    bot.BotHealth
	media     bot.BotMedia

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

		info:      facade,
		messaging: facade,
		threads:   facade,
		crypto:    facade,
		rooms:     facade,
		typing:    facade,
		sync:      facade,
		client:    facade,

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

// func (b *DefaultBot) AddEventHandler(h handlers.EventHandler) {
// 	b.dispatcher.AddEventHandler(h)
// }

// func (b *DefaultBot) AddFilter(f filters.Filter) {
// 	b.dispatcher.AddFilter(f)
// }

// func (b *DefaultBot) StartListening(ctx context.Context) error {
// 	return b.sync.StartListening(ctx)
// }

// func (b *DefaultBot) StopListening(ctx context.Context) error {
// 	return b.sync.StopListening(ctx)
// }

// func (b *DefaultBot) Name() string {
// 	return b.name
// }

// func (b *DefaultBot) FullName() string {
// 	return b.name // или @user:hs если хочешь
// }

// func (b *DefaultBot) Info() bot.BotInfo           { return b.info }
// func (b *DefaultBot) Messaging() bot.BotMessaging { return b.messaging }
// func (b *DefaultBot) Threads() bot.BotThreads     { return b.threads }
