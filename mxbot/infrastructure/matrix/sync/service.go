package sync

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/store"
)

var _ dbot.BotSync = (*Service)(nil)

type Service struct {
	client *mautrix.Client

	sink    dbot.EventSink
	auth    dbot.BotAuth
	deduper dbot.EventDeduper
	retry   time.Duration

	joinStore  *store.JoinStore
	prevBatch  *store.PrevBatchStore
	patchStart time.Time

	enableBackfill      bool
	backfillLimitPerReq int
	backfillDone        chan struct{}
	backfillCloseOnce   sync.Once

	startOnce sync.Once
	runOnce   sync.Once

	workCh      chan *event.Event
	numWorkers  int
	inflightTTL time.Duration

	cancel context.CancelFunc
}

func New(c dbot.BotClient, sink dbot.EventSink, retry, inflightTTL time.Duration, numWorkers int, opts ...Option) *Service {
	s := &Service{
		client: c.RawClient().(*mautrix.Client),
		sink:   sink,
		retry:  retry,

		backfillDone: make(chan struct{}),

		prevBatch: store.NewPrevBatchStore(),

		workCh:      make(chan *event.Event, 10000),
		numWorkers:  numWorkers,
		inflightTTL: inflightTTL,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Service) StartListening(ctx context.Context) error {
	var startErr error
	s.startOnce.Do(func() {
		startErr = s.startListening(ctx)
	})
	return startErr
}

func (s *Service) startListening(ctx context.Context) error {
	slog.Info("sync: StartListening", "user_id", s.client.UserID)

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Ensure syncer exists
	if s.client.Syncer == nil {
		s.client.Syncer = mautrix.NewDefaultSyncer()
	}

	ds, ok := s.client.Syncer.(*mautrix.DefaultSyncer)
	if !ok {
		return fmt.Errorf("unsupported syncer type: %T (need *mautrix.DefaultSyncer)", s.client.Syncer)
	}

	// Increase timeline limit (initial sync tail size)
	ds.FilterJSON = &mautrix.Filter{
		Room: &mautrix.RoomFilter{
			Timeline: &mautrix.FilterPart{
				Limit: 500,
			},
		},
	}

	svcCtx := ctx
	// We need prev_batch for backfill. It is only available in /sync responses,
	// so we capture it via OnSync.
	var backfillOnce sync.Once
	ds.OnSync(func(ctxSync context.Context, resp *mautrix.RespSync, since string) bool {
		totalTimeline := 0
		for _, roomData := range resp.Rooms.Join {
			totalTimeline += len(roomData.Timeline.Events)
		}
		slog.Info("sync: batch",
			"since", since,
			"next_batch", resp.NextBatch,
			"timeline_events", totalTimeline,
		)

		totalState := 0
		for _, roomData := range resp.Rooms.Join {
			totalTimeline += len(roomData.Timeline.Events)
			totalState += len(roomData.State.Events)
		}

		// 1) Save prev_batch tokens per room for /messages backfill
		for roomID, roomData := range resp.Rooms.Join {
			if roomData.Timeline.PrevBatch != "" && s.prevBatch != nil {
				s.prevBatch.Set(roomID, roomData.Timeline.PrevBatch)
			}
		}

		// 2) Save join timestamp for the bot user (room membership join)
		// We store it ourselves because StateEvent() in mautrix v0.24.0 returns only error (content only),
		// and doesn't give us origin_server_ts.
		if s.joinStore != nil {
			for roomID, roomData := range resp.Rooms.Join {
				events := append(roomData.State.Events, roomData.Timeline.Events...)
				for _, evt := range events {
					if evt.Type == event.StateMember && evt.GetStateKey() == s.client.UserID.String() {
						if membership, _ := evt.Content.Raw["membership"].(string); membership == "join" {
							_ = s.joinStore.SetIfLater(roomID, evt.Timestamp)
						}
					}
				}
			}
		}

		// 3) Start backfill once (after we have at least some prev_batch tokens)
		// if s.enableBackfill {
		backfillOnce.Do(func() {
			go s.backfillAllRooms(svcCtx)
		})
		// }

		return true
	})

	for i := 0; i < s.numWorkers; i++ {
		go s.worker(ctx)
	}

	ds.OnEvent(func(ctxEvt context.Context, evt *event.Event) {
		if evt.Type != event.EventMessage && evt.Type != event.EventEncrypted {
			return
		}

		// patchStart
		if !s.patchStart.IsZero() && evt.Timestamp > 0 &&
			time.UnixMilli(evt.Timestamp).Before(s.patchStart) {
			return
		}

		// dedup
		if s.deduper != nil && evt.ID != "" {
			ok, err := s.deduper.TryStartProcessing(ctx, evt.ID.String(), s.inflightTTL)
			if err != nil {
				slog.Error("dedup: TryStartProcessing failed", "err", err, "id", evt.ID)
				return
			}
			if !ok {
				return // already processed or already inflight
			}
		}

		// enqueue (dont block sync)
		select {
		case s.workCh <- evt:
			slog.Info("sync: got msg", "room", evt.RoomID, "id", evt.ID, "ts", evt.Timestamp)
		default:
			// queue is full: better remove inflight so we can try again
			slog.Error("sync: queue full, dropping", "room", evt.RoomID, "id", evt.ID)
			if s.deduper != nil && evt.ID != "" {
				_ = s.deduper.UnmarkInflight(ctx, evt.ID.String())
			}
		}
	})

	// Start sync loop
	s.runOnce.Do(func() {
		go s.run(ctx)
	})

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
		if err != nil {
			fmt.Printf("SYNC ERROR: %T %v\n", err, err)
		}

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

func (s *Service) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-s.workCh:
			if err := s.sink.HandleMatrixEvent(ctx, evt); err != nil {
				slog.Error("worker: HandleMatrixEvent failed", "err", err, "id", evt.ID, "room", evt.RoomID)
				if s.deduper != nil && evt.ID != "" {
					_ = s.deduper.UnmarkInflight(ctx, evt.ID.String())
				}
				continue
			}

			if s.deduper != nil && evt.ID != "" {
				if err := s.deduper.MarkProcessed(ctx, evt.ID.String()); err != nil {
					// processed, but couldn't record - possible replay, that's ok
					slog.Error("dedup: MarkProcessed failed", "err", err, "id", evt.ID)
				}
			}

			slog.Info("sync: handled ok", "room", evt.RoomID, "id", evt.ID, "ts", evt.Timestamp)
		}
	}
}
