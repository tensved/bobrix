package constructor

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"maunium.net/go/mautrix"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	dctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	dfilters "github.com/tensved/bobrix/mxbot/domain/filters"
	dhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"

	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/auth"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/client"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/crypto"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/ctx"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/events"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/health"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/info"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/media"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/messaging"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/rooms"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/sync"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/threads"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/typing"

	// applbot "github.com/tensved/bobrix/mxbot/application/bot"
	applctx "github.com/tensved/bobrix/mxbot/application/ctx"
	appldisp "github.com/tensved/bobrix/mxbot/application/dispatcher"
	applfilters "github.com/tensved/bobrix/mxbot/application/filters"
)

var (
	defaultSyncerRetryTime = 5 * time.Second
	defaultTypingTimeout   = 30 * time.Second
)

type Config struct {
	Credentials   *config.BotCredentials
	Logger        *zerolog.Logger
	TypingTimeout time.Duration
	SyncTimeout   time.Duration
	// Opts          applbot.BotOptions
}

type MatrixBot struct {
	dbot.BotAuth
	dbot.BotInfo
	dbot.BotMessaging
	dbot.BotThreads
	dbot.BotCrypto
	dbot.BotClient
	dbot.EventLoader
	dbot.BotRoomActions
	dbot.BotTyping
	dbot.BotSync
	dbot.BotHealth
	dbot.BotPresenceControl
	dbot.BotMedia

	dctx.CtxFactory
	Dispatcher *appldisp.Dispatcher
}

func NewMatrixBot(cfg Config) (*MatrixBot, error) {
	// --- raw Matrix client (no auth yet)
	clientProvider, err := client.New(cfg.Credentials.HomeServerURL)
	if err != nil {
		return nil, err
	}
	rawClient := clientProvider.RawClient().(*mautrix.Client)

	// --- authorize
	authSvc := auth.New(rawClient, cfg.Credentials, cfg.Credentials.Username)
	if err := authSvc.Authorize(context.Background()); err != nil {
		return nil, err
	}

	// --- crypto
	cryptoSvc, err := crypto.New(rawClient, cfg.Credentials.PickleKey, cfg.Credentials.Username)
	if err != nil {
		return nil, err
	}

	// --- rooms / threads / typing / messaging
	if cfg.TypingTimeout == 0 {
		cfg.TypingTimeout = defaultTypingTimeout
	}

	typingSvc := typing.New(clientProvider, cfg.TypingTimeout, cfg.Logger)
	roomsSvc := rooms.New(clientProvider)
	threadsSvc := threads.New(clientProvider, cfg.Credentials.IsThreadEnabled, cfg.Credentials.ThreadLimit)
	messagingSvc := messaging.New(clientProvider, cryptoSvc)
	infoSvc := info.New(clientProvider, cfg.Credentials.Username)

	// --- application ctx factory
	ctxFactory := applctx.NewFactory(
		messagingSvc,
		threadsSvc,
		nil, // event loader (пока не нужен)
		ctx.NewBotCtx(clientProvider, infoSvc),
	)

	// --- dispatcher (application)
	dispatcherSvc := appldisp.New(
		nil, // bot.FullBot get lower
		ctxFactory,
		[]dhandlers.EventHandler{}, // handlers get from application
		[]dfilters.Filter{
			applfilters.FilterNotMe(infoSvc),
			applfilters.FilterAfterStart(
				infoSvc,
				roomsSvc,
				applfilters.FilterAfterStartOptions{
					StartTime:      time.Now(),
					ProcessInvites: true,
				},
			),
		},
		cfg.Logger,
	)

	// --- events (decrypt → dispatch)
	eventsSvc := events.New(cryptoSvc, dispatcherSvc, dispatcherSvc.Filters())

	// --- sync (Matrix → events)
	syncSvc := sync.New(clientProvider, eventsSvc)

	healthSvc := health.New(clientProvider)

	mediaSvc := media.New(clientProvider)

	// --- final bot facade
	matrixBot := &MatrixBot{
		BotAuth:        authSvc,
		BotInfo:        infoSvc,
		BotMessaging:   messagingSvc,
		BotThreads:     threadsSvc,
		BotCrypto:      cryptoSvc,
		BotClient:      clientProvider,
		BotRoomActions: roomsSvc,
		BotTyping:      typingSvc,
		BotSync:        syncSvc,
		BotHealth:      healthSvc,
		BotMedia:       mediaSvc,
		CtxFactory:     ctxFactory,
		Dispatcher:     dispatcherSvc,
	}

	// inject FullBot into dispatcher
	dispatcherSvc.SetBot(matrixBot)

	return matrixBot, nil
}

func (b *MatrixBot) AddEventHandler(h dhandlers.EventHandler) {
	b.Dispatcher.AddEventHandler(h)
}

func (b *MatrixBot) AddFilter(f dfilters.Filter) {
	b.Dispatcher.AddFilter(f)
}
