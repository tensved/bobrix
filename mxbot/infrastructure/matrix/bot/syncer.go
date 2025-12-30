package bot // nok

import (
	"context"
	"fmt"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *DefaultBot) startSyncer(ctx context.Context) error {
	syncer := b.matrixClient.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEvent(func(eventCtx context.Context, evt *event.Event) {
		go b.eventHandler(eventCtx, evt)
	})

	// Encrypted Message Handler
	syncer.OnEventType(event.EventEncrypted, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Msg("received encrypted message")

		_, err := b.machine.DecryptMegolmEvent(ctx, evt)
		if err != nil {
			b.logger.Error().
				Err(err).
				Str("room_id", evt.RoomID.String()).
				Str("sender", string(evt.Sender)).
				Msg("failed to decrypt message")

			if err.Error() == "no session with given ID found" {
				// Request the key from the sender
				err = b.machine.SendRoomKeyRequest(ctx, evt.RoomID, evt.Content.AsEncrypted().SenderKey, evt.Content.AsEncrypted().SessionID, "m.megolm.v1", map[id.UserID][]id.DeviceID{
					evt.Sender: {id.DeviceID(evt.Content.AsEncrypted().DeviceID)},
				})
				if err != nil {
					b.logger.Error().Err(err).Msg("failed to request room key")
					return
				}
				b.logger.Info().Msg("sent room key request")
			}
			return
		}

		b.logger.Info().Msg("message decrypted successfully")
	})

	// Room Key Receiver Handler
	syncer.OnEventType(event.ToDeviceRoomKey, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Interface("content", evt.Content.Raw).Msg("received room key event")
		content := evt.Content.AsRoomKey()
		if content == nil {
			b.logger.Error().Msg("invalid room key format")
			return
		}

		b.machine.HandleToDeviceEvent(ctx, evt)
	})

	// Key Request Handler
	syncer.OnEventType(event.ToDeviceRoomKeyRequest, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Msg("received room key request")
		content := evt.Content.AsRoomKeyRequest()
		if content == nil {
			b.logger.Error().Msg("invalid room key request format")
			return
		}

		b.logger.Info().Str("sender", string(evt.Sender)).Msg("room key request from")
		b.machine.HandleToDeviceEvent(ctx, evt)
	})

	// Handler for receiving forwarded room keys
	syncer.OnEventType(event.ToDeviceForwardedRoomKey, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Msg("received forwarded room key")
		content := evt.Content.AsForwardedRoomKey()
		if content == nil {
			b.logger.Error().Msg("invalid forwarded room key format")
			return
		}

		b.machine.HandleToDeviceEvent(ctx, evt)
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				b.logger.Info().Msg("syncer stopped by context")
				return

			default:
				b.logger.Info().Msg("start sync")
				err := b.matrixClient.SyncWithContext(ctx)

				if ctx.Err() != nil {
					return
				}

				if err != nil {
					httpErr, ok := err.(mautrix.HTTPError)

					if ok && httpErr.RespError.StatusCode == 401 {
						b.logger.Warn().Msg("token expired, reauth needed")

						if err := b.authBot(ctx); err != nil {
							b.logger.Error().Err(err).Msg("failed to reauth bot")
							time.Sleep(b.syncerTimeRetry)
							continue
						}

						b.logger.Info().Msg("reauth success, restarting sync")
						continue
					}

					b.logger.Error().Err(err).Msg("failed to sync")
					time.Sleep(b.syncerTimeRetry)
				}
			}
		}
	}()

	return nil
}

// prepareBot - Authenticates the bot with the homeserver
// and register the bot if it is not registered
// also refreshes the access token if it is expired
func (b *DefaultBot) prepareBot(ctx context.Context) error {
	if err := b.authorizeBot(ctx); err != nil {
		return err
	}

	if b.displayName != "" {
		if err := b.matrixClient.SetDisplayName(ctx, b.displayName); err != nil {
			return err
		}
	}

	// Initialize crypto after successful login
	if err := b.initCrypto(ctx); err != nil {
		return fmt.Errorf("failed to init crypto: %w", err)
	}

	return nil
}
