package typing

import (
	"context"
	"sync"
	"time"

	"maunium.net/go/mautrix/id"
)

type roomTyping struct {
	mu sync.Mutex

	timer *time.Timer

	// loop lifecycle
	running bool
	refs    int            // number of active callers holding a stop func
	cancel  func()         // idempotent stop
	done    <-chan struct{} // closed when loop ends
}

// EnsureTyping guarantees a maximum of one typing loop per roomID (within a single bot).
// Each call extends the TTL fallback and returns a stop func.
// Calling the returned stop func decrements the ref-count; the loop is cancelled
// only when the last caller stops (or the TTL fallback fires).
func (b *Service) EnsureTyping(workerCtx context.Context, roomID id.RoomID, ttl time.Duration) func() {
	if ttl <= 0 {
		ttl = b.typingTimeout
	}

	noop := func() {}

	v, _ := b.rooms.LoadOrStore(roomID, &roomTyping{})
	rt := v.(*roomTyping)

	// --- short section under per-room lock: extend TTL and decide whether to start loop
	var needStart bool
	var loopCtx context.Context
	var loopCancel context.CancelFunc

	rt.mu.Lock()

	// (re)arm TTL fallback timer
	if rt.timer == nil {
		rt.timer = time.NewTimer(ttl)
	} else {
		if !rt.timer.Stop() {
			select {
			case <-rt.timer.C:
			default:
			}
		}
		rt.timer.Reset(ttl)
	}

	if !rt.running {
		rt.running = true
		needStart = true
		loopCtx, loopCancel = context.WithCancel(workerCtx)
		rt.cancel = func() { loopCancel() }
	}

	rt.refs++
	refs := rt.refs
	timer := rt.timer
	rt.mu.Unlock()

	_ = refs // used implicitly via the closure below

	// Build a stop func that decrements the ref-count and cancels only when last.
	var once sync.Once
	stopFunc := func() {
		once.Do(func() {
			rt.mu.Lock()
			rt.refs--
			remaining := rt.refs
			cancel := rt.cancel
			rt.mu.Unlock()

			if remaining <= 0 && cancel != nil {
				cancel()
			}
		})
	}

	// --- if loop already exists, just extended TTL and return a stop func
	if !needStart {
		return stopFunc
	}

	cancelLoop, done, err := b.LoopTyping(loopCtx, roomID)
	if err != nil {
		loopCancel()

		rt.mu.Lock()
		rt.running = false
		rt.refs = 0
		rt.cancel = nil
		rt.done = nil
		if rt.timer != nil {
			rt.timer.Stop()
			rt.timer = nil
		}
		rt.mu.Unlock()

		b.rooms.Delete(roomID)
		return noop
	}

	stop := func() {
		loopCancel()
		cancelLoop()
	}

	rt.mu.Lock()
	rt.cancel = stop
	rt.done = done
	rt.mu.Unlock()

	go b.watch(roomID, rt, timer, done)

	return stopFunc
}

func (b *Service) watch(roomID id.RoomID, rt *roomTyping, timer *time.Timer, done <-chan struct{}) {
	select {
	case <-timer.C:
		rt.mu.Lock()
		if !rt.running {
			rt.mu.Unlock()
			return
		}
		rt.running = false
		rt.refs = 0
		cancel := rt.cancel
		rt.cancel = nil
		rt.done = nil
		if rt.timer == timer {
			rt.timer = nil
		}
		rt.mu.Unlock()

		if cancel != nil {
			cancel()
		}

	case <-done:
		rt.mu.Lock()
		if !rt.running {
			rt.mu.Unlock()
			return
		}
		rt.running = false
		rt.refs = 0
		rt.cancel = nil
		rt.done = nil
		if rt.timer != nil {
			rt.timer.Stop()
			rt.timer = nil
		}
		rt.mu.Unlock()
	}

	b.rooms.Delete(roomID)
}
