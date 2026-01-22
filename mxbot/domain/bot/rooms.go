package bot

import (
	"context"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type BotRoomActions interface {
	JoinRoom(ctx context.Context, roomID id.RoomID) error
	JoinedMembersCount(ctx context.Context, roomID id.RoomID) (int, error)
	
	GetJoinedRoomsList(ctx context.Context) ([]id.RoomID, error)
	GetMessagesFromRoomByNumber(ctx context.Context, roomID id.RoomID, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error)
	GetMessagesFromRoomByDuration(ctx context.Context, roomID id.RoomID, duration time.Duration, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error)
}
