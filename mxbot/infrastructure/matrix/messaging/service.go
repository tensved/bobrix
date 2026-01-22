package messaging

import (
	"context"
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/messages"
)

var _ dbot.BotMessaging = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	crypto dbot.BotCrypto
}

func New(c dbot.BotClient, crypto dbot.BotCrypto) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
		crypto: crypto,
	}
}

func (s *Service) SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error {
	if msg == nil {
		return dbot.ErrNilMessage
	}

	if msg.Type().IsMedia() {
		resp, err := s.client.UploadMedia(ctx, msg.AsReqUpload())
		if err != nil {
			return fmt.Errorf("%w: %w", dbot.ErrUploadMedia, err)
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
