package client

import (
	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
)

var _ domainbot.BotClient = (*Provider)(nil)

type Provider struct {
	client *mautrix.Client
}

func New(cfg Config) (*Provider, error) {
	cli, err := mautrix.NewClient(
		cfg.HomeserverURL,
		cfg.UserID,
		cfg.AccessToken,
	)
	if err != nil {
		return nil, err
	}

	return &Provider{
		client: cli,
	}, nil
}

func (p *Provider) RawClient() any {
	return p.client
}
