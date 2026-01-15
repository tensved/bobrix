package bot

import "context"

type BotHealth interface {
	Ping(ctx context.Context) error
	SetOnlineStatus()
	SetOfflineStatus()
	SetIdleStatus()
}
