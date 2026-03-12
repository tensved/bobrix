package typing

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ dombot.BotTyping = (*Service)(nil)

type Service struct {
	client        *mautrix.Client
	typingTimeout time.Duration
	logger        *zerolog.Logger

	baseCtx context.Context

	// key: id.RoomID, value: *roomTyping
	rooms sync.Map
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
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	return &Service{
		client:        c.RawClient().(*mautrix.Client),
		typingTimeout: typingTimeout,
		logger:        logger,
		baseCtx:       baseCtx,
	}
}

func (b *Service) GetTypingTimeout() time.Duration { return b.typingTimeout }
