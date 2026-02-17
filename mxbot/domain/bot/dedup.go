package bot

import (
	"context"
	"time"
)

// Dedyplication
// EventDeduper returns true if event was not processed before and is now marked as processed.
type EventDeduper interface {
	// TryStartProcessing sets the inflight lease to TTL if:
	// - the event hasn't been processed yet
	// - and isn't inflight (or the inflight has expired)
	// Returns ok=true if we've captured the event and need to process it.
	TryStartProcessing(ctx context.Context, eventID string, ttl time.Duration) (ok bool, err error)

	// MarkProcessed marks successful processing (translates to processed, removes inflight).
	MarkProcessed(ctx context.Context, eventID string) error

	// UnmarkInflight removes inflight (for example, if processing has failed or the queue is full).
	UnmarkInflight(ctx context.Context, eventID string) error

	// IsProcessed is needed for the backfill stop condition and simply for diagnostics.
	IsProcessed(ctx context.Context, eventID string) (bool, error)
}
