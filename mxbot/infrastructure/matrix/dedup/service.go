package dedup

import (
	"context"
	"sync"
	"time"

	bot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ bot.EventDeduper = (*LeaseDeduper)(nil)

type LeaseDeduper struct {
	mu sync.Mutex

	processed map[string]struct{}
	inflight  map[string]time.Time // eventID -> expiresAt

	// optional GC
	gcEvery time.Duration
	stopGC  chan struct{}
}

func NewLeaseDeduper(gcEvery time.Duration) *LeaseDeduper {
	if gcEvery <= 0 {
		gcEvery = 30 * time.Second
	}
	d := &LeaseDeduper{
		processed: make(map[string]struct{}),
		inflight:  make(map[string]time.Time),
		gcEvery:   gcEvery,
		stopGC:    make(chan struct{}),
	}
	go d.gcLoop()
	return d
}

func (d *LeaseDeduper) Close() {
	close(d.stopGC)
}

func (d *LeaseDeduper) gcLoop() {
	t := time.NewTicker(d.gcEvery)
	defer t.Stop()

	for {
		select {
		case <-d.stopGC:
			return
		case now := <-t.C:
			d.mu.Lock()
			for id, exp := range d.inflight {
				if !exp.IsZero() && now.After(exp) {
					delete(d.inflight, id)
				}
			}
			d.mu.Unlock()
		}
	}
}

func (d *LeaseDeduper) TryStartProcessing(_ context.Context, eventID string, ttl time.Duration) (bool, error) {
	if eventID == "" {
		return true, nil
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	now := time.Now()
	exp := now.Add(ttl)

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.processed[eventID]; ok {
		return false, nil
	}

	if curExp, ok := d.inflight[eventID]; ok {
		// если lease ещё жив — не даём второй раз
		if curExp.After(now) {
			return false, nil
		}
		// lease истёк — можно перезахватить
	}

	d.inflight[eventID] = exp
	return true, nil
}

func (d *LeaseDeduper) MarkProcessed(_ context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	d.processed[eventID] = struct{}{}
	delete(d.inflight, eventID)
	return nil
}

func (d *LeaseDeduper) UnmarkInflight(_ context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.inflight, eventID)
	return nil
}

func (d *LeaseDeduper) IsProcessed(_ context.Context, eventID string) (bool, error) {
	if eventID == "" {
		return false, nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.processed[eventID]
	return ok, nil
}

// package dedup

// import (
// 	"context"
// 	"sync"

// 	"github.com/tensved/bobrix/mxbot/domain/bot"
// )

// var _ bot.EventDeduper = (*MemoryDeduper)(nil)

// type MemoryDeduper struct {
// 	mu sync.Mutex
// 	m  map[string]struct{}
// }

// func NewMemoryDeduper() *MemoryDeduper {
// 	return &MemoryDeduper{m: map[string]struct{}{}}
// }

// func (d *MemoryDeduper) TryMarkProcessed(_ context.Context, eventID string) bool {
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
// 	if _, ok := d.m[eventID]; ok {
// 		return false
// 	}
// 	d.m[eventID] = struct{}{}
// 	return true
// }

// func (d *MemoryDeduper) MarkProcessed(_ context.Context, eventID string) error {
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
// 	d.m[eventID] = struct{}{}
// 	return nil
// }

// func (d *MemoryDeduper) IsProcessed(_ context.Context, eventID string) (bool, error) {
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
// 	_, ok := d.m[eventID]
// 	return ok, nil
// }
