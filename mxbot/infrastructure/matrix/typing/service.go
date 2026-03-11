package typing

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ dombot.BotTyping = (*Service)(nil)

type Service struct {
	client        *mautrix.Client
	typingTimeout time.Duration
	logger        *zerolog.Logger

	baseCtx context.Context

	typingMu sync.Mutex
	typing   map[id.RoomID]*typingState
}

func New(
	c dombot.BotClient,
	typingTimeout time.Duration,
	logger *zerolog.Logger,
	baseCtx context.Context,
) *Service {
	if typingTimeout <= 0 {
		typingTimeout = 5 * time.Second
	}
	if logger == nil {
		l := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &l
	}
	return &Service{
		client:        c.RawClient().(*mautrix.Client),
		typingTimeout: typingTimeout,
		logger:        logger,

		baseCtx: baseCtx,

		typingMu: sync.Mutex{},
		typing:   map[id.RoomID]*typingState{},
	}
}

func (b *Service) GetTypingTimeout() time.Duration {
	return b.typingTimeout
}
