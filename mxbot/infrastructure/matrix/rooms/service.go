package rooms

import (
	"context"
	"fmt"

	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
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
		return fmt.Errorf("%w roomID=%v: %w", ErrJoinToRoom, roomID, err)
	}

	return nil
}
