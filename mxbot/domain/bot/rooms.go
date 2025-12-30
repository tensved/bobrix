package bot

import (
	"context"

	"maunium.net/go/mautrix/id"
)

type BotRoomActions interface {
	JoinRoom(ctx context.Context, roomID id.RoomID) error
	JoinedMembersCount(ctx context.Context, roomID id.RoomID) (int, error)
}
