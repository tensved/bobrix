package events

import (
	"context"

	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
)

var _ bot.EventRouter = (*Service)(nil)

type Service struct {
	crypto  bot.BotCrypto
	sink    bot.EventSink
	filters []filters.Filter
}

func New(
	crypto bot.BotCrypto,
	sink bot.EventSink,
	filters []filters.Filter,
) *Service {
	return &Service{
		crypto:  crypto,
		sink:    sink,
		filters: filters,
	}
}

func (s *Service) HandleMatrixEvent(ctx context.Context, evt *event.Event) error {
	// --- 1. to-device events (crypto-level)
	switch evt.Type {
	case event.ToDeviceRoomKey,
		event.ToDeviceRoomKeyRequest,
		event.ToDeviceForwardedRoomKey:
		s.crypto.HandleToDevice(ctx, evt)
		return nil
	}

	if evt.RoomID == "" {
		s.crypto.HandleToDevice(ctx, evt)
		return nil
	}

	// --- 2. decrypt if needed
	decrypted, err := s.crypto.DecryptEvent(ctx, evt)
	if err != nil {
		return err
	}

	// --- 3. transport-level filters (mxbot)
	for _, f := range s.filters {
		if !f(decrypted) {
			return nil // silently drop
		}
	}

	// --- 4. pass to application
	return s.sink.HandleMatrixEvent(ctx, decrypted)
}
