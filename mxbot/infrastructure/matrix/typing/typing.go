package typing

import (
	"context"
	"sync"
	"time"

	"maunium.net/go/mautrix/id"
)

// typingData - struct for store information about typing by bots in rooms
// it is used to avoid problem with cancelling typingEvent when bot have multiple requests at the same time
type typingData struct {
	data map[string]int
	mx   *sync.RWMutex
}

func (d *typingData) add(roomID id.RoomID) {
	d.mx.Lock()
	d.data[roomID.String()]++
	d.mx.Unlock()
}

func (d *typingData) remove(roomID id.RoomID) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.data[roomID.String()]--

	if d.data[roomID.String()] == 0 {
		delete(d.data, roomID.String())
	}
}

var typing = typingData{
	data: make(map[string]int),
	mx:   &sync.RWMutex{},
}

// LoopTyping - Starts and stops typing on the room. Typing is sent every typingTimeout
// it will be stopped if the context is cancelled or if an error occurs
// it returns a function that can be used to stop the typing
func (b *Service) LoopTyping(ctx context.Context, roomID id.RoomID) (cancelTyping func(), err error) {

	ticker := time.NewTicker(b.typingTimeout)

	typing.add(roomID)

	err = b.StartTyping(ctx, roomID)
	if err != nil {
		return nil, err
	}

	cancel := make(chan struct{})

	go func(c <-chan struct{}) {
		for {
			select {
			case <-ticker.C:
				if err := b.StartTyping(ctx, roomID); err != nil {
					b.logger.Error().Err(err).Msg("failed to stop typing")
				}

			case <-c:
				typing.remove(roomID)

				if err := b.StopTyping(ctx, roomID); err != nil {
					b.logger.Error().Err(err).Msg("failed to stop typing")
				}

				return
			}
		}
	}(cancel)

	return func() {
		cancel <- struct{}{}
		ticker.Stop()

		close(cancel)
	}, nil

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
