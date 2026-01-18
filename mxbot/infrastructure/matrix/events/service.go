package events

import (
	"context"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix/event"
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


// package events

// import (
// 	"context"

// 	"github.com/tensved/bobrix/mxbot/domain/bot"
// 	domain "github.com/tensved/bobrix/mxbot/domain/bot"
// 	"maunium.net/go/mautrix"
// 	"maunium.net/go/mautrix/event"
// 	"maunium.net/go/mautrix/id"
// )

// type Service struct {
// 	client *mautrix.Client
// 	crypto domain.BotCrypto
// 	router domain.EventRouter
// }

// func New(
// 	c domain.BotClient,
// 	crypto domain.BotCrypto,
// 	router domain.EventRouter,
// ) *Service {
// 	return &Service{
// 		client: c.RawClient().(*mautrix.Client),
// 		crypto: crypto,
// 		router: router,
// 	}
// }

// var _ domain.EventLoader = (*Service)(nil)

// func (s *Service) GetEvent(ctx context.Context, roomID id.RoomID, eventID id.EventID) (*event.Event, error) {
// 	evt, err := s.client.GetEvent(ctx, roomID, eventID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return s.crypto.DecryptEvent(ctx, evt)
// }

// var _ bot.EventSink = (*Service)(nil)

// func (s *Service) HandleMatrixEvent(ctx context.Context, evt *event.Event) error {
// 	decrypted, err := s.crypto.DecryptEvent(ctx, evt)
// 	if err != nil {
// 		return err
// 	}

// 	return s.router.HandleMatrixEvent(ctx, decrypted)
// }
