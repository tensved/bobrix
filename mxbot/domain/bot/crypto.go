package bot

import (
	"context"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type BotCrypto interface {
	Encrypt(ctx context.Context, roomID id.RoomID, eventType event.Type, content any) (event.Type, any, error)
	DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error)

	IsEncrypted(evt *event.Event) bool
	IsEncryptedRoom(ctx context.Context, roomID id.RoomID) (bool, error)

	EnsureOutboundSession(ctx context.Context, roomID id.RoomID) error
	RequestKey(ctx context.Context, evt *event.Event) error
	HandleToDevice(ctx context.Context, evt *event.Event)
}
