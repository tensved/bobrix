package sync

import (
	"time"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/store"
)

type Option func(*Service)

func WithAuth(a dbot.BotAuth) Option {
	return func(s *Service) {
		s.auth = a
	}
}

func WithDeduper(d dbot.EventDeduper) Option {
	return func(s *Service) {
		s.deduper = d
	}
}

func WithPatchStart(t time.Time) Option {
	return func(s *Service) {
		s.patchStart = t
	}
}

// func WithBackfill(enabled bool) Option {
// 	return func(s *Service) {
// 		s.enableBackfill = enabled
// 	}
// }

func WithBackfillLimitPerRequest(n int) Option {
	return func(s *Service) {
		if n > 0 {
			s.backfillLimitPerReq = n
		}
	}
}

func WithJoinStore(js *store.JoinStore) Option {
	return func(s *Service) {
		s.joinStore = js
	}
}

func WithRetry(d time.Duration) Option {
	return func(s *Service) {
		if d > 0 {
			s.retry = d
		}
	}
}

func WithWorkers(w int) Option {
	return func(s *Service) {
		s.numWorkers = w
	}
}

func WithInflightTTL(d time.Duration) Option {
	return func(s *Service) {
		s.inflightTTL = d
	}
}
