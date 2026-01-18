package bot

import "maunium.net/go/mautrix/event"

type BotPresenceControl interface {
	SetStatus(status event.Presence)
	SetIdleStatus()
	SetOnlineStatus()
	SetOfflineStatus()
}
