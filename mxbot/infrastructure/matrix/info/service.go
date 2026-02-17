package info

import (
	"errors"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ dbot.BotInfo = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	name   string
}

func New(c dbot.BotClient, name string) (*Service, error) {
	if name == "" {
		return nil, errors.New("bot name shouldnt be an empty string")
	}
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		name:   name,
	}, nil
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
