package info

import (
	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

var _ domain.BotInfo = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	name   string
}

func New(c domain.BotClient, name string) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		name:   name,
	}
}

func (b *Service) Name() string {
	return b.name
}

func (b *Service) FullName() string {
	return b.client.UserID.String()
}

func (b *Service) UserID() id.UserID {
	return b.client.UserID
}
