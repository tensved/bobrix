package health

import (
	"context"

	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
)

var _ domain.BotHealth = (*Service)(nil)

type Service struct {
	client *mautrix.Client
}

func New(c domain.BotClient) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
	}
}

// ????????????????
// Ping - Checks if the bot is online
// It will return error if the bot is offline
func (s *Service) Ping(ctx context.Context) error {
	_, err := s.client.GetOwnDisplayName(ctx)
	return err
}

func (s *Service) SetOnlineStatus() {
	// _, err := s.client.GetOwnDisplayName(ctx)
}

func (s *Service) SetOfflineStatus() {
	// _, err := s.client.GetOwnDisplayName(ctx)

}

func (s *Service) SetIdleStatus() {
	// _, err := s.client.GetOwnDisplayName(ctx)

}
