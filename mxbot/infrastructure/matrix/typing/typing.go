package typing

import (
	"context"
	"errors"
	"sync"
	"time"

	"maunium.net/go/mautrix/id"
)

type typingState struct {
	cancel context.CancelFunc
	timer  *time.Timer
	done   <-chan struct{} // Closes when the loop actually finishes
}

// EnsureTyping guarantees a maximum of one typing loop per roomID.
// Each new call extends the TTL. When the TTL expires, typing is disabled and the state is removed from the map.
// If the loop terminates early (ctx is canceled/error occurs), the state is also immediately removed from the map.
func (b *Service) EnsureTyping(workerCtx context.Context, roomID id.RoomID, ttl time.Duration) {
	b.typingMu.Lock()
	if b.typing == nil {
		b.typing = make(map[id.RoomID]*typingState)
	}

	// prolongation
	if st, ok := b.typing[roomID]; ok {
		st.timer.Stop()
		st.timer.Reset(ttl)
		b.typingMu.Unlock()
		return
	}

	loopCtx, loopCancel := context.WithCancel(workerCtx)

	cancelLoop, done, err := b.LoopTyping(loopCtx, roomID)
	if err != nil {
		loopCancel()
		b.typingMu.Unlock()
		return
	}

	stop := func() {
		// stop goroutine
		loopCancel()
		// terminate loop (idempotent)
		cancelLoop()
	}

	st := &typingState{cancel: stop, done: done}
	st.timer = time.AfterFunc(ttl, func() {
		b.typingMu.Lock()
		cur, ok := b.typing[roomID]
		if ok && cur == st {
			delete(b.typing, roomID)
		}
		b.typingMu.Unlock()

		if ok && cur == st {
			st.cancel()
		}
	})

	b.typing[roomID] = st
	b.typingMu.Unlock()

	// clear the map immediately if the loop terminates before the ttl (shutdown/error)
	go func() {
		<-done
		b.typingMu.Lock()
		cur, ok := b.typing[roomID]
		if ok && cur == st {
			delete(b.typing, roomID)
			st.timer.Stop()
		}
		b.typingMu.Unlock()
	}()
}

// LoopTyping starts typing=true immediately and then extends it every typingTimeout.
// Returns:
// - cancelTyping: stop the loop (idempotent)
// - done: closes when the loop has completed
func (b *Service) LoopTyping(loopCtx context.Context, roomID id.RoomID) (cancel func(), done <-chan struct{}, err error) {
	ticker := time.NewTicker(b.typingTimeout)

	// helper: each typing request with a separate timeout
	doStart := func() error {
		reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return b.StartTyping(reqCtx, roomID)
	}
	doStop := func() error {
		reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return b.StopTyping(reqCtx, roomID)
	}

	if err := doStart(); err != nil {
		ticker.Stop()
		return nil, nil, err
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	var once sync.Once

	go func() {
		defer ticker.Stop()
		defer close(doneCh)

		// on any exit we try to turn off typing
		defer func() {
			if err := doStop(); err != nil {
				// context canceled will almost never happen here, but just in case:
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					b.logger.Error().Err(err).Msg("failed to stop typing")
				}
			}
		}()

		for {
			select {
			case <-ticker.C:
				if err := doStart(); err != nil {
					if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
						b.logger.Error().Err(err).Msg("failed to start typing")
					}
				}

			case <-loopCtx.Done():
				return

			case <-stopCh:
				return
			}
		}
	}()

	return func() { once.Do(func() { close(stopCh) }) }, doneCh, nil
}

// StartTyping - Starts typing on the room
func (b *Service) StartTyping(ctx context.Context, roomID id.RoomID) error {
	_, err := b.client.UserTyping(ctx, roomID, true, b.typingTimeout)
	return err
}

// StopTyping - Stops typing on the room
func (b *Service) StopTyping(ctx context.Context, roomID id.RoomID) error {
	_, err := b.client.UserTyping(ctx, roomID, false, b.typingTimeout)
	return err
}
