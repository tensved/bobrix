package rooms

import (
	"context"
	"fmt"
	"slices"
	"time"

	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var _ domain.BotRoomActions = (*Service)(nil)

type Service struct {
	client *mautrix.Client
}

func New(c domain.BotClient) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
	}
}

func (s *Service) JoinedMembersCount(ctx context.Context, roomID id.RoomID) (int, error) {
	resp, err := s.client.JoinedMembers(ctx, roomID)
	if err != nil {
		return 0, err
	}
	return len(resp.Joined), nil
}

func (s *Service) JoinRoom(ctx context.Context, roomID id.RoomID) error {
	_, err := s.client.JoinRoomByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("%w roomID=%v: %w", domain.ErrJoinToRoom, roomID, err)
	}

	return nil
}

func (s *Service) GetJoinedRoomsList(ctx context.Context) ([]id.RoomID, error) {
	rooms, err := s.client.JoinedRooms(ctx)
	return rooms.JoinedRooms, err
}

func (s *Service) GetMessagesFromRoomByNumber(ctx context.Context, roomID id.RoomID, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error) {
	resp, err := s.client.Messages(ctx, roomID, "", "", 'b', filter, numMessages)
	if err != nil {
		return []*event.Event{}, err
	}

	messages := make([]*event.Event, 0, len(resp.Chunk))
	for i := len(resp.Chunk) - 1; i >= 0; i-- { // evts ascending order by the time
		messages = append(messages, resp.Chunk[i])

	}
	return messages, nil
}

func (s *Service) GetMessagesFromRoomByDuration(ctx context.Context, roomID id.RoomID, duration time.Duration, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error) {
	timeAgo := time.Now().Add(-1 * duration).UnixMilli()

	syncResp, err := s.client.SyncRequest(ctx, 0, "", "", false, "")
	if err != nil {
		return nil, err
	}
	from := syncResp.NextBatch

	var allMessages []*event.Event

	for {
		msgsResp, err := s.client.Messages(ctx, roomID, from, "", 'b', filter, numMessages)
		if err != nil {
			return nil, err
		}

		stop := false
		for _, evt := range msgsResp.Chunk {
			if evt.Timestamp <= timeAgo {
				stop = true
				break
			}
			allMessages = append(allMessages, evt)
		}

		if stop || msgsResp.End == "" {
			break
		}

		from = msgsResp.End
	}

	slices.Reverse(allMessages) // evts ascending order by the time

	return allMessages, nil
}
