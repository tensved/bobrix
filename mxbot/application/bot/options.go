package bot

import "time"

//??????????????
var (
	defaultSyncerRetryTime = 5 * time.Second
	defaultTypingTimeout   = 30 * time.Second
)

// mxbot/application/bot/options.go
// Bot options. Used to configure the bot
type BotOptions func(*DefaultBot)

// WithSyncerRetryTime - time to wait before retrying a failed sync
func WithSyncerRetryTime(d time.Duration) BotOptions {
	return func(bot *DefaultBot) {
		bot.syncerTimeRetry = d
	}
}

// WithTypingTimeout - time to wait before sending a typing event
func WithTypingTimeout(d time.Duration) BotOptions {
	return func(bot *DefaultBot) {
		bot.typingTimeout = d
	}
}

func WithDisplayName(name string) BotOptions {
	return func(bot *DefaultBot) {
		bot.displayName = name
	}
}
