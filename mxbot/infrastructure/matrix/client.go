package bot

import "maunium.net/go/mautrix"

type BotClient interface {
	Client() *mautrix.Client
}
