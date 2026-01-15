package messaging

import (
	"context"
	"fmt"

	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var _ domain.BotMessaging = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	crypto domain.BotCrypto
}

func New(c domain.BotClient, crypto domain.BotCrypto) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		crypto: crypto,
	}
}

func (s *Service) SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error {
	if msg == nil {
		return domain.ErrNilMessage
	}

	if msg.Type().IsMedia() {
		resp, err := s.client.UploadMedia(ctx, msg.AsReqUpload())
		if err != nil {
			return fmt.Errorf("%w: %w", domain.ErrUploadMedia, err)
		}
		msg.SetContentURI(resp.ContentURI)
	}

	encrypted, err := s.crypto.IsEncryptedRoom(ctx, roomID)
	if err != nil {
		return err
	}

	if !encrypted {
		_, err := s.client.SendMessageEvent(ctx, roomID, event.EventMessage, msg.AsJSON())
		return err
	}

	if err := s.crypto.EnsureOutboundSession(ctx, roomID); err != nil {
		return err
	}

	evType, content, err := s.crypto.Encrypt(ctx, roomID, event.EventMessage, msg.AsJSON())
	if err != nil {
		return err
	}

	_, err = s.client.SendMessageEvent(ctx, roomID, evType, content)
	return err
}

// ?????????????????
// func (s *Service) SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error {
// 	if msg == nil {
// 		return domain.ErrNilMessage
// 	}

// 	if msg.Type().IsMedia() {
// 		uploadResponse, err := s.client.UploadMedia(ctx, msg.AsReqUpload())
// 		if err != nil {
// 			return fmt.Errorf("%w: %w", domain.ErrUploadMedia, err)
// 		}
// 		msg.SetContentURI(uploadResponse.ContentURI)
// 	}

// 	// Checking if the room is encrypted
// 	var encryptionEvent event.EncryptionEventContent
// 	err := s.client.StateEvent(ctx, roomID, event.StateEncryption, "", &encryptionEvent)
// 	if err != nil {
// 		// If get 404, it means the room is not encrypted.
// 		if err.Error() == "M_NOT_FOUND (HTTP 404): Event not found." {
// 			// Send a not encrypted message
// 			_, err = s.client.SendMessageEvent(ctx, roomID, event.EventMessage, msg.AsJSON())
// 			if err != nil {
// 				return fmt.Errorf("%w: %w", domain.ErrSendMessage, err)
// 			}
// 			return nil
// 		}
// 		slog.Error("failed to get room encryption state", "error", err)
// 		return fmt.Errorf("%w: %w", domain.ErrSendMessage, err)
// 	}

// 	// Let's check if we have a session for this room
// 	session, err := s.machine.CryptoStore.GetOutboundGroupSession(ctx, roomID)
// 	if err != nil || session == nil {
// 		// If there is no session, create a new one
// 		// Get a list of room participants
// 		members, err := s.client.Members(ctx, roomID)
// 		if err != nil {
// 			slog.Error("failed to get room members", "error", err)
// 			return fmt.Errorf("%w: %w", domain.ErrSendMessage, err)
// 		}

// 		// Collecting a list of user IDs
// 		userIDs := make([]id.UserID, 0, len(members.Chunk))
// 		for _, member := range members.Chunk {
// 			if member.StateKey != nil {
// 				userIDs = append(userIDs, id.UserID(*member.StateKey))
// 			}
// 		}

// 		err = s.machine.ShareGroupSession(ctx, roomID, userIDs)
// 		if err != nil {
// 			return fmt.Errorf("%w: %w", domain.ErrSendMessage, err)
// 		}
// 	}

// 	// Determine event type based on message type
// 	eventType := event.EventMessage
// 	if msg.Type().IsMedia() {
// 		switch msg.Type() {
// 		case event.MsgImage:
// 			eventType = event.EventMessage
// 		case event.MsgVideo:
// 			eventType = event.EventMessage
// 		case event.MsgAudio:
// 			eventType = event.EventMessage
// 		case event.MsgFile:
// 			eventType = event.EventMessage
// 		}
// 	}

// 	encryptedContent, err := s.machine.EncryptMegolmEvent(ctx, roomID, eventType, msg.AsJSON())
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", domain.ErrSendMessage, err)
// 	}

// 	// If encryption was successful, send the encrypted message
// 	_, err = s.client.SendMessageEvent(ctx, roomID, event.EventEncrypted, encryptedContent)
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", domain.ErrSendMessage, err)
// 	}

// 	return nil
// }
