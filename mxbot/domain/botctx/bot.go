package botctx

import (
	"maunium.net/go/mautrix/id"
)

type Bot interface {
	UserID() id.UserID
	Name() string
	FullName() string
	RawClient() any
}
