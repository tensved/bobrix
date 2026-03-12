package typing

import (
	"context"
	"errors"
	"sync"
	"time"

	"maunium.net/go/mautrix/id"
)

// type typingState struct {
// 	cancel context.CancelFunc
// 	timer  *time.Timer
// 	done   <-chan struct{} // Closes when the loop actually finishes
// }

// LoopTyping starts typing=true immediately and then extends it every typingTimeout.
// Returns:
// - cancelTyping: stop the loop (idempotent)
// - done: closes when the loop has completed
func (b *Service) LoopTyping(loopCtx context.Context, roomID id.RoomID) (cancel func(), done <-chan struct{}, err error) {
	ticker := time.NewTicker(b.typingTimeout)

	doStart := func() error {
		reqCtx, cancel := context.WithTimeout(b.baseCtx, 5*time.Second)
		defer cancel()
		return b.StartTyping(reqCtx, roomID)
	}
	doStop := func() error {
		reqCtx, cancel := context.WithTimeout(b.baseCtx, 5*time.Second)
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

		defer func() {
			if err := doStop(); err != nil {
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
