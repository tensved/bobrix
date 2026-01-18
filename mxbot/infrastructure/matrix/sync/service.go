package sync

import (
	"context"
	"time"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var _ bot.BotSync = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	router bot.EventRouter

	auth  bot.Auth
	sink  bot.EventSink
	retry time.Duration

	cancel context.CancelFunc
}

func New(c bot.BotClient, router bot.EventRouter) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		router: router,
	}
}

// ?????????????????
func (s *Service) StartListening(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	syncer := s.client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		_ = s.sink.HandleMatrixEvent(ctx, evt)
	})

	go s.run(ctx)
	return nil
}

func (s *Service) StopListening(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func (s *Service) run(ctx context.Context) {
	for {
		err := s.client.SyncWithContext(ctx)

		if ctx.Err() != nil {
			return
		}

		if httpErr, ok := err.(mautrix.HTTPError); ok &&
			httpErr.RespError.StatusCode == 401 {

			if err := s.auth.Reauthorize(ctx); err != nil {
				time.Sleep(s.retry)
				continue
			}
			continue
		}

		time.Sleep(s.retry)
	}
}
