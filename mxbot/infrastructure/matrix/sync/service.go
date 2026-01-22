package sync

import (
	"context"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ dbot.BotSync = (*Service)(nil)

type Service struct {
	client *mautrix.Client

	sink  dbot.EventSink
	auth  dbot.BotAuth
	retry time.Duration

	cancel context.CancelFunc
}

func New(c dbot.BotClient, sink dbot.EventSink) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		sink:   sink,
	}
}

func (s *Service) StartListening(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	if s.client.Syncer == nil {
		panic("SYNCER IS NIL")
	}

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

			if err := s.auth.Authorize(ctx); err != nil {
				time.Sleep(s.retry)
				continue
			}
			continue
		}

		time.Sleep(s.retry)
	}
}
