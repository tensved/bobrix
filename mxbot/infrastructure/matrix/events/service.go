package events

import (
	"context"

	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ bot.EventRouter = (*Service)(nil)

type Service struct {
	crypto bot.BotCrypto
	sink   bot.EventSink
}

func New(crypto bot.BotCrypto, sink bot.EventSink) *Service {
	return &Service{
		crypto: crypto,
		sink:   sink,
	}
}

func (s *Service) HandleMatrixEvent(ctx context.Context, evt *event.Event) error {
	// --- 1. to-device events (crypto-level)
	switch evt.Type {
	case event.ToDeviceRoomKey,
		event.ToDeviceRoomKeyRequest,
		event.ToDeviceForwardedRoomKey,
		event.EventEncrypted:
		s.crypto.HandleToDevice(ctx, evt)
		return nil
	}

	if evt.RoomID == "" {
		s.crypto.HandleToDevice(ctx, evt)
		return nil
	}

	// --- 2. encrypted → decrypt
	decrypted, err := s.crypto.DecryptEvent(ctx, evt)
	if err != nil {
		return err
	}

	return s.sink.HandleMatrixEvent(ctx, decrypted)
}
