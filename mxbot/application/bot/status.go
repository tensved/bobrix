package bot

import (
	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ bot.BotPresenceControl = (*DefaultBot)(nil)

func (b *DefaultBot) SetStatus(status event.Presence) {
	b.botStatus = status
}

func (b *DefaultBot) SetIdleStatus() {
	b.SetStatus(event.PresenceUnavailable)
}

func (b *DefaultBot) SetOnlineStatus() {
	b.SetStatus(event.PresenceOnline)
}

func (b *DefaultBot) SetOfflineStatus() {
	b.SetStatus(event.PresenceOffline)
}
