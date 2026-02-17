package typing

import (
	"os"
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
}

func New(
	c dombot.BotClient,
	typingTimeout time.Duration,
	logger *zerolog.Logger,
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
	}
}
