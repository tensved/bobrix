package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var _ bot.BotSync = (*Service)(nil)

type Service struct {
	client *mautrix.Client

	sink  bot.EventSink
	auth  bot.BotAuth
	retry time.Duration

	cancel context.CancelFunc
}

func New(c bot.BotClient, sink bot.EventSink) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		sink:   sink,
	}
}

func (s *Service) StartListening(ctx context.Context) error {
	slog.Info("SYNC: Start listening")
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	if s.client.Syncer == nil {
		panic("SYNCER IS NIL")
	}

	slog.Info("SYNC: syncer", "type", fmt.Sprintf("%T", s.client.Syncer))

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
		slog.Info("SYNC: calling SyncWithContext")
		err := s.client.SyncWithContext(ctx)

		if ctx.Err() != nil {
			slog.Info("SYNC: Ctx done")
			return
		}

		if httpErr, ok := err.(mautrix.HTTPError); ok &&
			httpErr.RespError.StatusCode == 401 {

			if err := s.auth.Authorize(ctx); err != nil { //???
				slog.Error("SYNC: sync error", "err", err)
				time.Sleep(s.retry)
				continue
			}
			continue
		}

		time.Sleep(s.retry)
	}
}
