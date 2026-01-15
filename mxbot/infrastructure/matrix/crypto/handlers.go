package crypto

import (
	"context"

	"maunium.net/go/mautrix/event"
)

func (s *Service) HandleToDevice(ctx context.Context, evt *event.Event) {
	s.machine.HandleToDeviceEvent(ctx, evt)
}

func (s *Service) RequestKey(ctx context.Context, evt *event.Event) error {
	c := evt.Content.AsEncrypted()
	return s.machine.SendRoomKeyRequest(
		ctx,
		evt.RoomID,
		c.SenderKey,
		c.SessionID,
		"m.megolm.v1",
		nil,
	)
}

