package bot

import (
	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix/event"
)

var _ dbot.BotPresenceControl = (*DefaultBot)(nil)

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
