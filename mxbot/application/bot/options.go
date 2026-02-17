package bot

// Bot options. Used to configure the bot
type BotOptions func(*DefaultBot)

func WithDisplayName(name string) BotOptions {
	return func(bot *DefaultBot) {
		bot.displayName = name
	}
}
