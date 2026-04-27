package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	utils "github.com/tensved/bobrix/mxbot/infrastructure/utils"
)

type Service struct {
	client  *mautrix.Client
	machine *crypto.OlmMachine

	encMu   sync.RWMutex
	encRoom map[id.RoomID]bool
}

var _ dbot.BotCrypto = (*Service)(nil)

func New(client *mautrix.Client, pickleKey []byte, name string) (*Service, error) {
	dir := filepath.Join(".bin", "crypto")
	_ = os.MkdirAll(dir, 0700)

	safeUser := utils.SafeFilePart(name)

	store := filepath.Join(dir, "store-"+safeUser+".db")

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
		encRoom: make(map[id.RoomID]bool),
	}, nil
}

func (s *Service) IsEncrypted(evt *event.Event) bool {
	return evt != nil && evt.Type == event.EventEncrypted
}

func (s *Service) DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error) {
	if evt.Type != event.EventEncrypted {
		return evt, nil
	}

	decrypted, err := s.machine.DecryptMegolmEvent(ctx, evt)
	if err == nil {
		return decrypted, nil
	}

	if err.Error() == "no session with given ID found" {
		_ = s.RequestKey(ctx, evt)
		return nil, err
	}

	return nil, err
}

// IsEncryptedRoom now does NOT make remote requests.
// Returns:
// - (true, nil) if we know the room is encrypted
// - (false, nil) if we know the room is NOT encrypted (if you'll be marking this at some point)
// - (false, err) if the state is unknown (cold start / state not seen)
func (s *Service) IsEncryptedRoom(_ context.Context, roomID id.RoomID) (bool, error) {
	if s == nil {
		return false, fmt.Errorf("crypto service is nil")
	}

	s.encMu.RLock()
	v, ok := s.encRoom[roomID]
	s.encMu.RUnlock()

	if !ok {
		return false, fmt.Errorf("encryption state unknown for room %s", roomID)
	}

	return v, nil
}

func (s *Service) MarkRoomEncrypted(roomID id.RoomID) {
	if s == nil || roomID == "" {
		return
	}
	s.encMu.Lock()
	s.encRoom[roomID] = true
	s.encMu.Unlock()
}

func (s *Service) MarkRoomNotEncrypted(roomID id.RoomID) {
	if s == nil || roomID == "" {
		return
	}
	s.encMu.Lock()
	s.encRoom[roomID] = false
	s.encMu.Unlock()
}

func (s *Service) ObserveEvent(evt *event.Event) {
	if s == nil || evt == nil {
		return
	}

	if evt.Type == event.StateEncryption {
		s.MarkRoomEncrypted(evt.RoomID)
		return
	}

	if evt.Type == event.EventEncrypted {
		s.MarkRoomEncrypted(evt.RoomID)
		return
	}

	s.MarkRoomNotEncrypted(evt.RoomID)
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

func (s *Service) Encrypt(ctx context.Context, roomID id.RoomID, evType event.Type, content any) (event.Type, any, error) {
	encrypted, err := s.machine.EncryptMegolmEvent(ctx, roomID, evType, content)
	if err != nil {
		return event.Type{}, nil, err
	}
	return event.EventEncrypted, encrypted, nil
}
