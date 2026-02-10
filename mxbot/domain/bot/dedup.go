package bot

import (
	"context"
	"time"
)

// Dedyplication
// EventDeduper returns true if event was not processed before and is now marked as processed.
type EventDeduper interface {
	// TryStartProcessing ставит inflight lease на ttl, если:
	// - event ещё не processed
	// - и не inflight (или inflight истёк)
	// Возвращает ok=true если мы "захватили" событие и должны его обработать.
	TryStartProcessing(ctx context.Context, eventID string, ttl time.Duration) (ok bool, err error)

	// MarkProcessed фиксирует успешную обработку (переводит в processed, снимает inflight).
	MarkProcessed(ctx context.Context, eventID string) error

	// UnmarkInflight снимает inflight (например, если обработка упала или очередь переполнена).
	UnmarkInflight(ctx context.Context, eventID string) error

	// IsProcessed нужен для backfill stop condition и просто для диагностики.
	IsProcessed(ctx context.Context, eventID string) (bool, error)
}
