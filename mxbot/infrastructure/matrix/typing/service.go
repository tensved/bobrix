package typing

import (
	"time"

	"github.com/rs/zerolog"
	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
)

var _ domain.BotTyping = (*Service)(nil)

type Service struct {
	client        *mautrix.Client
	typingTimeout time.Duration
	logger        *zerolog.Logger
}

func New(
	c domain.BotClient,
	typingTimeout time.Duration,
	logger *zerolog.Logger,
) *Service {
	return &Service{
		client:        c.RawClient().(*mautrix.Client),
		typingTimeout: typingTimeout,
		logger:        logger,
	}
}
