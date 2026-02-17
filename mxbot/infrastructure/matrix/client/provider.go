package client

import (
	"maunium.net/go/mautrix"

	"github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ bot.BotClient = (*Provider)(nil)

type Provider struct {
	client *mautrix.Client
}

func New(hsURL string, store mautrix.SyncStore) (*Provider, error) {
	cli, err := mautrix.NewClient(hsURL, "", "")
	if err != nil {
		return nil, err
	}

	cli.Store = store

	if cli.Syncer == nil {
		cli.Syncer = mautrix.NewDefaultSyncer()
	}

	return &Provider{client: cli}, nil
}

func (p *Provider) RawClient() any { return p.client }
