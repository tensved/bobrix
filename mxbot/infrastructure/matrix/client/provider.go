package client

import (
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
)

var _ bot.BotClient = (*Provider)(nil)

type Provider struct {
	client *mautrix.Client
}

func New(HomeserverURL string) (*Provider, error) {
	cli, err := mautrix.NewClient(HomeserverURL, "", "")
	if err != nil {
		return nil, err
	}

	return &Provider{client: cli}, nil
}

func (p *Provider) RawClient() any {
	return p.client
}
