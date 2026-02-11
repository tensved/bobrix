package dedup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	bot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/infrastructure/repository/pg"
)

var _ bot.EventDeduper = (*PostgresDeduper)(nil)

const (
	statusInflight  int16 = 1
	statusProcessed int16 = 2
)

// ---- processed cache ----
// Cache only "processed" (final state).
type processedCache struct {
	mu  sync.Mutex
	ttl time.Duration
	max int
	m   map[string]time.Time // eventID -> expiresAt
}

func newProcessedCache(ttl time.Duration, max int) *processedCache {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if max <= 0 {
		max = 200_000
	}
	return &processedCache{
		ttl: ttl,
		max: max,
		m:   make(map[string]time.Time, max),
	}
}

func (c *processedCache) Has(eventID string) bool {
	if eventID == "" {
		return false
	}
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	exp, ok := c.m[eventID]
	if !ok {
		return false
	}
	if now.After(exp) {
		delete(c.m, eventID)
		return false
	}
	return true
}

func (c *processedCache) Put(eventID string) {
	if eventID == "" {
		return
	}
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	// simple growth preventer
	if len(c.m) >= c.max {
		c.m = make(map[string]time.Time, c.max)
	}
	c.m[eventID] = now.Add(c.ttl)
}

// ---- Deduper ----

type PostgresDeduper struct {
	provider pg.ExecutorProvider
	cache    *processedCache
}

type PostgresDeduperOptions struct {
	ProcessedCacheTTL time.Duration
	ProcessedCacheMax int
}

func NewPostgresDeduper(provider pg.ExecutorProvider, opts PostgresDeduperOptions) *PostgresDeduper {
	return &PostgresDeduper{
		provider: provider,
		cache:    newProcessedCache(opts.ProcessedCacheTTL, opts.ProcessedCacheMax),
	}
}

func (d *PostgresDeduper) TryStartProcessing(ctx context.Context, eventID string, ttl time.Duration) (bool, error) {
	if eventID == "" {
		return true, nil
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	// fast failure: recently processed
	if d.cache != nil && d.cache.Has(eventID) {
		return false, nil
	}

	exec := d.provider.Get(ctx)

	q := `
		INSERT INTO matrix_event_dedup(event_id, status, lease_until, processed_at, updated_at)
		VALUES ($1, $2, now() + ($3 * interval '1 second'), NULL, now())
		ON CONFLICT (event_id) DO UPDATE
		SET status = EXCLUDED.status,
			lease_until = EXCLUDED.lease_until,
			updated_at = now()
		WHERE
		matrix_event_dedup.status <> $4
		AND (matrix_event_dedup.lease_until IS NULL OR matrix_event_dedup.lease_until < now())
		RETURNING event_id
		`

	sec := int64(ttl.Seconds())
	if sec <= 0 {
		sec = 1
	}

	var returned string
	err := exec.QueryRow(ctx, q, eventID, statusInflight, sec, statusProcessed).Scan(&returned)
	if err == nil {
		return true, nil
	}
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (d *PostgresDeduper) MarkProcessed(ctx context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}

	exec := d.provider.Get(ctx)

	q := `
		INSERT INTO matrix_event_dedup(event_id, status, lease_until, processed_at, updated_at)
		VALUES ($1, $2, NULL, now(), now())
		ON CONFLICT (event_id) DO UPDATE
		SET status = $2,
			lease_until = NULL,
			processed_at = now(),
			updated_at = now()
		`
	_, err := exec.Exec(ctx, q, eventID, statusProcessed)
	if err == nil && d.cache != nil {
		d.cache.Put(eventID)
	}
	return err
}

func (d *PostgresDeduper) UnmarkInflight(ctx context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}

	exec := d.provider.Get(ctx)

	q := `
		UPDATE matrix_event_dedup
		SET lease_until = NULL,
			updated_at = now()
		WHERE event_id = $1 AND status <> $2
		`
	_, err := exec.Exec(ctx, q, eventID, statusProcessed)
	return err
}

func (d *PostgresDeduper) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	if eventID == "" {
		return false, nil
	}

	if d.cache != nil && d.cache.Has(eventID) {
		return true, nil
	}

	exec := d.provider.Get(ctx)

	var status int16
	err := exec.QueryRow(ctx,
		`SELECT status FROM matrix_event_dedup WHERE event_id=$1`,
		eventID,
	).Scan(&status)

	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("dedup IsProcessed query failed: %w", err)
	}

	ok := status == statusProcessed
	if ok && d.cache != nil {
		d.cache.Put(eventID)
	}
	return ok, nil
}
