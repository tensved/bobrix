package bot // ok2

import (
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix/id"
)

var _ bot.Info = (*DefaultBot)(nil)

func (b *DefaultBot) Name() string {
	return b.name
}

func (b *DefaultBot) FullName() string {
	return b.client.UserID.String()
}

func (b *DefaultBot) UserID() id.UserID {
	return b.client.UserID
}
