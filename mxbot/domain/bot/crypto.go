package bot

import (
	"context"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type BotCrypto interface {
	IsEncryptedRoom(ctx context.Context, roomID id.RoomID) (bool, error)

	EnsureOutboundSession(ctx context.Context, roomID id.RoomID) error

	Encrypt(
		ctx context.Context,
		roomID id.RoomID,
		eventType event.Type,
		content any,
	) (event.Type, any, error)

	IsEncrypted(evt *event.Event) bool
	DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error)
	RequestKey(ctx context.Context, evt *event.Event) error
	HandleToDevice(ctx context.Context, evt *event.Event)
}
