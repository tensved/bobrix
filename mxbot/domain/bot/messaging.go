package bot // ok

import (
	"context"

	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/id"
)

type BotMessaging interface {
	SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error
}
