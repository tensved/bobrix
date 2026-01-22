package bot

import "maunium.net/go/mautrix/id"

// Info provides read-only information about the bot identity.
type BotInfo interface {
	UserID() id.UserID
	FullName() string // get full name with servername (e.g. @username:servername)
	Name() string
}
