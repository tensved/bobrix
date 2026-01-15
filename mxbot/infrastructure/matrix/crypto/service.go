package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Service struct {
	client  *mautrix.Client
	machine *crypto.OlmMachine
}

var _ domain.BotCrypto = (*Service)(nil)

func New(client *mautrix.Client, pickleKey []byte, name string) (*Service, error) {
	dir := filepath.Join(".bin", "crypto")
	os.MkdirAll(dir, 0700)

	store := filepath.Join(dir, "store-"+name+".db")

	helper, err := cryptohelper.NewCryptoHelper(client, pickleKey, store)
	if err != nil {
		return nil, err
	}

	if err := helper.Init(context.Background()); err != nil {
		return nil, err
	}

	m := helper.Machine()
	if err := m.Load(context.Background()); err != nil {
		return nil, err
	}

	return &Service{
		client:  client,
		machine: m,
	}, nil
}

func (s *Service) IsEncrypted(evt *event.Event) bool {
	return evt.Type == event.EventEncrypted
}

func (s *Service) DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error) {
	if !s.IsEncrypted(evt) {
		return evt, nil
	}
	return s.machine.DecryptMegolmEvent(ctx, evt)
}

func (s *Service) IsEncryptedRoom(ctx context.Context, roomID id.RoomID) (bool, error) {
	var enc event.EncryptionEventContent

	err := s.client.StateEvent(ctx, roomID, event.StateEncryption, "", &enc)
	if err != nil {
		if httpErr, ok := err.(mautrix.HTTPError); ok && httpErr.RespError.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to fetch encryption state: %w", err)
	}

	return true, nil
}

func (s *Service) EnsureOutboundSession(ctx context.Context, roomID id.RoomID) error {
	session, err := s.machine.CryptoStore.GetOutboundGroupSession(ctx, roomID)
	if err == nil && session != nil {
		return nil
	}

	members, err := s.client.Members(ctx, roomID)
	if err != nil {
		return err
	}

	var users []id.UserID
	for _, m := range members.Chunk {
		if m.StateKey != nil {
			users = append(users, id.UserID(*m.StateKey))
		}
	}

	return s.machine.ShareGroupSession(ctx, roomID, users)
}

func (s *Service) Encrypt(
	ctx context.Context,
	roomID id.RoomID,
	evType event.Type,
	content any,
) (event.Type, any, error) {

	encrypted, err := s.machine.EncryptMegolmEvent(ctx, roomID, evType, content)
	if err != nil {
		return event.Type{}, nil, err
	}

	return event.EventEncrypted, encrypted, nil
}
