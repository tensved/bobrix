package sync

import (
	"context"
	"log/slog"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (s *Service) backfillRoom(ctx context.Context, roomID id.RoomID) error {
	// Backfill only makes sense if it is enabled and there is a deduper
	// (otherwise, "exactly once" cannot be ensured)
	if !s.enableBackfill || s.deduper == nil {
		return nil
	}

	// 1) pagination start token
	from, ok := s.prevBatch.Get(roomID)
	if !ok || from == "" {
		// prev_batch hasn't synced yet — backfill isn't possible yet
		return nil
	}

	// 2) time limit: max(last_join_ts, patchStart) (patchStart is optional)
	start := s.patchStart

	var joinTime time.Time
	if s.joinStore != nil {
		if joinTS, ok := s.joinStore.Get(roomID); ok && joinTS > 0 {
			joinTime = time.UnixMilli(joinTS)
		}
	}

	if start.IsZero() {
		start = joinTime
	} else if !joinTime.IsZero() && joinTime.After(start) {
		start = joinTime
	}

	// if start is still zero, don't backfill
	// (unsafe: we'll end up in an infinite loop)
	if start.IsZero() {
		return nil
	}

	for {
		// Messages(ctx, roomID, from, to, dir, filter *FilterPart, limit int)
		resp, err := s.client.Messages(ctx, roomID, from, "", mautrix.DirectionBackward, nil, s.backfillLimitPerReq)
		if err != nil {
			return err
		}
		if resp == nil || len(resp.Chunk) == 0 {
			return nil
		}

		// resp.Chunk goes from new to old
		// we unfold it to process from old to new
		for i := len(resp.Chunk) - 1; i >= 0; i-- {
			evt := resp.Chunk[i]

			if evt.Type != event.EventMessage && evt.Type != event.EventEncrypted {
				continue
			}

			if evt.Timestamp > 0 && time.UnixMilli(evt.Timestamp).Before(start) {
				return nil
			}

			if s.deduper != nil && evt.ID != "" {
				processed, err := s.deduper.IsProcessed(ctx, evt.ID.String())
				if err != nil {
					return err
				}
				if processed {
					continue
				}
			}

			if err := s.eventRouter.HandleMatrixEvent(ctx, evt); err != nil {
				// don't mark processed so we can repeat
				slog.Error("backfill: HandleMatrixEvent failed",
					"err", err, "type", evt.Type.String(), "room", evt.RoomID, "id", evt.ID)
				continue
			}

			if s.deduper != nil && evt.ID != "" {
				if err := s.deduper.MarkProcessed(ctx, evt.ID.String()); err != nil {
					return err
				}
			}
		}

		if resp.End == "" || resp.End == from {
			return nil
		}
		from = resp.End
	}
}

func (s *Service) backfillAllRooms(ctx context.Context) {
	defer s.backfillCloseOnce.Do(
		func() {
			close(s.backfillDone)
		},
	)

	joined, err := s.client.JoinedRooms(ctx)
	if err != nil {
		return
	}

	for _, roomID := range joined.JoinedRooms {
		if err := s.backfillRoom(ctx, roomID); err != nil {
			slog.Error("backfillRoom", "roomID", roomID, "err", err)
		}
	}
}
