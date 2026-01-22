package typing

import (
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
	return &Service{
		client:        c.RawClient().(*mautrix.Client),
		typingTimeout: typingTimeout,
		logger:        logger,
	}
}
