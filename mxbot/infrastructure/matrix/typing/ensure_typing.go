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
	cancel  func()          // idempotent stop
	done    <-chan struct{} // closed when loop ends
}

// EnsureTyping guarantees a maximum of one typing loop per roomID (within a single bot).
// Each call extends the TTL. After the TTL, typing is disabled and the entry is deleted.
func (b *Service) EnsureTyping(workerCtx context.Context, roomID id.RoomID, ttl time.Duration) {
	if ttl <= 0 {
		ttl = b.typingTimeout
	}

	v, _ := b.rooms.LoadOrStore(roomID, &roomTyping{})
	rt := v.(*roomTyping)

	// --- short section under per-room lock: extend TTL and decide whether to start loop
	var needStart bool
	var loopCtx context.Context
	var loopCancel context.CancelFunc

	rt.mu.Lock()

	// (re)arm timer
	if rt.timer == nil {
		rt.timer = time.NewTimer(ttl)
	} else {
		if !rt.timer.Stop() {
			// drain if fired
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

		// temporary cancel until we get a real cancelLoop from LoopTyping
		rt.cancel = func() { loopCancel() }
	}

	timer := rt.timer
	rt.mu.Unlock()

	// --- if loop already exists, just extend TTL
	if !needStart {
		return
	}

	cancelLoop, done, err := b.LoopTyping(loopCtx, roomID)
	if err != nil {
		loopCancel()

		// rollback + cleanup
		rt.mu.Lock()
		rt.running = false
		rt.cancel = nil
		rt.done = nil
		if rt.timer != nil {
			rt.timer.Stop()
			rt.timer = nil
		}
		rt.mu.Unlock()

		b.rooms.Delete(roomID)
		return
	}

	stop := func() {
		loopCancel()
		cancelLoop()
	}

	rt.mu.Lock()
	rt.cancel = stop
	rt.done = done
	rt.mu.Unlock()

	// watcher: TTL
	go b.watchTTL(roomID, rt, timer)

	// watcher: done
	go b.watchDone(roomID, rt, done)
}

func (b *Service) watchTTL(roomID id.RoomID, rt *roomTyping, timer *time.Timer) {
	<-timer.C

	rt.mu.Lock()
	if !rt.running {
		rt.mu.Unlock()
		return
	}
	rt.running = false
	cancel := rt.cancel
	rt.cancel = nil
	rt.done = nil
	// timer consumed
	if rt.timer == timer {
		rt.timer = nil
	}
	rt.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	b.rooms.Delete(roomID)
}

func (b *Service) watchDone(roomID id.RoomID, rt *roomTyping, done <-chan struct{}) {
	<-done

	rt.mu.Lock()
	if !rt.running {
		rt.mu.Unlock()
		return
	}
	rt.running = false
	rt.cancel = nil
	rt.done = nil
	if rt.timer != nil {
		rt.timer.Stop()
		rt.timer = nil
	}
	rt.mu.Unlock()

	b.rooms.Delete(roomID)
}
