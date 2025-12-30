package bot // nok

import (
	"context"
	"fmt"
	"log/slog"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var _ domainbot.BotMessaging = (*DefaultBot)(nil)

func (b *DefaultBot) SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error {
	if msg == nil {
		return ErrNilMessage
	}

	if msg.Type().IsMedia() {
		uploadResponse, err := b.matrixClient.UploadMedia(ctx, msg.AsReqUpload())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrUploadMedia, err)
		}
		msg.SetContentURI(uploadResponse.ContentURI)
	}

	// Checking if the room is encrypted
	var encryptionEvent event.EncryptionEventContent
	err := b.matrixClient.StateEvent(ctx, roomID, event.StateEncryption, "", &encryptionEvent)
	if err != nil {
		// If get 404, it means the room is not encrypted.
		if err.Error() == "M_NOT_FOUND (HTTP 404): Event not found." {
			// Send a not encrypted message
			_, err = b.matrixClient.SendMessageEvent(ctx, roomID, event.EventMessage, msg.AsJSON())
			if err != nil {
				return fmt.Errorf("%w: %w", ErrSendMessage, err)
			}
			return nil
		}
		slog.Error("failed to get room encryption state", "error", err)
		return fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	// Let's check if we have a session for this room
	session, err := b.machine.CryptoStore.GetOutboundGroupSession(ctx, roomID)
	if err != nil || session == nil {
		// If there is no session, create a new one
		// Get a list of room participants
		members, err := b.matrixClient.Members(ctx, roomID)
		if err != nil {
			slog.Error("failed to get room members", "error", err)
			return fmt.Errorf("%w: %w", ErrSendMessage, err)
		}

		// Collecting a list of user IDs
		userIDs := make([]id.UserID, 0, len(members.Chunk))
		for _, member := range members.Chunk {
			if member.StateKey != nil {
				userIDs = append(userIDs, id.UserID(*member.StateKey))
			}
		}

		err = b.machine.ShareGroupSession(ctx, roomID, userIDs)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrSendMessage, err)
		}
	}

	// Determine event type based on message type
	eventType := event.EventMessage
	if msg.Type().IsMedia() {
		switch msg.Type() {
		case event.MsgImage:
			eventType = event.EventMessage
		case event.MsgVideo:
			eventType = event.EventMessage
		case event.MsgAudio:
			eventType = event.EventMessage
		case event.MsgFile:
			eventType = event.EventMessage
		}
	}

	encryptedContent, err := b.machine.EncryptMegolmEvent(ctx, roomID, eventType, msg.AsJSON())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	// If encryption was successful, send the encrypted message
	_, err = b.matrixClient.SendMessageEvent(ctx, roomID, event.EventEncrypted, encryptedContent)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	return nil
}

func (b *DefaultBot) JoinRoom(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.JoinRoomByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("%w roomID=%v: %w", ErrJoinToRoom, roomID, err)
	}

	return nil
}

// Download - downloads the content of the mxc URL
func (b *DefaultBot) Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error) {
	return b.matrixClient.DownloadBytes(ctx, mxcURL)
}

// StartTyping - Starts typing on the room
func (b *DefaultBot) StartTyping(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.UserTyping(ctx, roomID, true, b.typingTimeout)
	return err
}

// StopTyping - Stops typing on the room
func (b *DefaultBot) StopTyping(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.UserTyping(ctx, roomID, false, b.typingTimeout)
	return err
}

// Ping - Checks if the bot is online
// It will return error if the bot is offline
func (b *DefaultBot) Ping(ctx context.Context) error {
	_, err := b.matrixClient.GetOwnDisplayName(ctx)
	return err
}
