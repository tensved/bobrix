package constructor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rs/zerolog"
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix"

	dctx "github.com/tensved/bobrix/mxbot/domain/ctx"
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
	"github.com/tensved/bobrix/mxbot/application/dispatcher"
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
	bot.BotAuth
	bot.BotInfo
	bot.BotMessaging
	bot.BotThreads
	bot.BotCrypto
	bot.BotClient
	bot.EventLoader
	bot.BotRoomActions
	bot.BotTyping
	bot.BotSync
	bot.BotHealth
	bot.BotPresenceControl
	bot.BotMedia

	dctx.CtxFactory
	Dispatcher *dispatcher.Dispatcher
}

func NewMatrixBot(cfg Config, mxbotFilters []filters.Filter) (*MatrixBot, error) {
	// --- raw Matrix client (no auth yet)
	clientProvider, err := client.New(cfg.Credentials.HomeServerURL)
	if err != nil {
		return nil, err
	}
	rawClient := clientProvider.RawClient().(*mautrix.Client)

	// --- authorize
	authSvc := auth.New(rawClient, cfg.Credentials, cfg.Credentials.Username)
	if err := authSvc.Authorize(context.Background()); err != nil {
		slog.Info("--------------", "auth err", err)
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
	dispatcherSvc := dispatcher.New(
		nil, // bot.FullBot будет присвоен ниже
		ctxFactory,
		[]dhandlers.EventHandler{}, // handlers передаются из application
		nil,                        // global filters
		cfg.Logger,
	)

	slog.Info("CTOR dispatcher", "ptr", fmt.Sprintf("%p", dispatcherSvc))

	// --- events (decrypt → dispatch)
	eventsSvc := events.New(cryptoSvc, dispatcherSvc, mxbotFilters)

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

func (b *MatrixBot) AddFilter(f filters.Filter) {
	b.Dispatcher.AddFilter(f)
}
