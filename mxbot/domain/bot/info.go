package bot // ok2

import "maunium.net/go/mautrix/id"

// Info provides read-only information about the bot identity.
type Info interface {
	UserID() id.UserID
	FullName() string
	Name() string
}
