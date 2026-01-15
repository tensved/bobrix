package sync

import (
	"context"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var _ bot.BotSync = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	router bot.EventRouter

	cancel context.CancelFunc
}

func New(c bot.BotClient, router bot.EventRouter) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		router: router,
	}
}

func (s *Service) StartListening(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	syncer := s.client.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		s.router.HandleMatrixEvent(ctx, evt)
	})

	go s.client.SyncWithContext(ctx)

	return nil
}

func (s *Service) StopListening(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// type Service struct {
// 	client *mautrix.Client
// 	crypto domain.BotCrypto
// 	router domain.EventRouter
// 	events domain.EventLoader
// 	sink   domain.EventSink
// }

// func New(
// 	c domain.BotClient,
// 	crypto domain.BotCrypto,
// 	router domain.EventRouter,
// 	events domain.EventLoader,
// 	sink domain.EventSink,
// ) *Service {
// 	return &Service{
// 		client: c.RawClient().(*mautrix.Client),
// 		crypto: crypto,
// 		router: router,
// 		events: events,
// 		sink:   sink,
// 	}
// }

// func (s *Service) Start(ctx context.Context) {
// 	syncer := s.client.Syncer.(*mautrix.DefaultSyncer)

// 	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
// 		s.sink.HandleMatrixEvent(ctx, evt)
// 	})

// 	syncer.OnEventType(event.ToDeviceRoomKey, s.crypto.HandleToDevice)
// 	syncer.OnEventType(event.ToDeviceRoomKeyRequest, s.crypto.HandleToDevice)
// 	syncer.OnEventType(event.ToDeviceForwardedRoomKey, s.crypto.HandleToDevice)

// 	go s.client.SyncWithContext(ctx)
// }
