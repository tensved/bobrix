package health

import (
	"context"

	"maunium.net/go/mautrix"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ dbot.BotHealth = (*Service)(nil)

type Service struct {
	client *mautrix.Client
}

func New(c dbot.BotClient) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
	}
}

// Ping - Checks if the bot is online
// It will return error if the bot is offline
func (s *Service) Ping(ctx context.Context) error {
	_, err := s.client.GetOwnDisplayName(ctx)
	return err
}
