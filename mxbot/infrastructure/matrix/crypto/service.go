package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	utils "github.com/tensved/bobrix/mxbot/infrastructure/utils"
)

// olmAccountMismatchMsgs are the error messages mautrix's cryptohelper returns
// from verifyDeviceKeysOnServer when the locally persisted olm account state
// disagrees with what the homeserver has on record for this device:
//   - the local account isn't marked as shared, but the server already has keys
//     for the device (e.g. the process was killed between uploading keys and
//     persisting the "shared" flag locally);
//   - the local account is marked as shared, but the server has no keys for the
//     device anymore (e.g. the device was deleted/logged out server-side);
//   - the identity key on the server for this device doesn't match the local
//     olm account's identity key (the local crypto store and the device id are
//     out of sync, e.g. one was reset/reused without the other).
// In all three cases the fix is the same: the local crypto store and device id
// are unrecoverably out of sync with the server, so reset both and log in with
// a fresh device.
var olmAccountMismatchMsgs = []string{
	"olm account is not marked as shared, but there are keys on the server",
	"olm account is marked as shared, keys seem to have disappeared from the server",
	"mismatching identity key on server",
}

// IsAccountKeyMismatch reports whether err is one of the local/remote olm
// account desyncs described above.
func IsAccountKeyMismatch(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, m := range olmAccountMismatchMsgs {
		if strings.Contains(msg, m) {
			return true
		}
	}
	return false
}

// ResetLocalState removes the persisted crypto store and saved device id for
// the given bot username, forcing a brand-new device and olm account to be
// created on the next login. Use this to recover from IsAccountKeyMismatch.
func ResetLocalState(name string) error {
	safeUser := utils.SafeFilePart(name)
	dir := filepath.Join(".bin", "crypto")

	matches, err := filepath.Glob(filepath.Join(dir, "store-"+safeUser+".db*"))
	if err != nil {
		return fmt.Errorf("glob crypto store: %w", err)
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", m, err)
		}
	}

	deviceIDFile := filepath.Join(dir, fmt.Sprintf("device-id-%s.txt", safeUser))
	if err := os.Remove(deviceIDFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", deviceIDFile, err)
	}

	return nil
}

type Service struct {
	client  *mautrix.Client
	machine *crypto.OlmMachine
	helper  *cryptohelper.CryptoHelper

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
		helper:  helper,
		encRoom: make(map[id.RoomID]bool),
	}, nil
}

// Close flushes and closes the underlying crypto store. It must be called
// during a graceful shutdown so the olm account's "shared" state is
// persisted to disk; skipping this is what allows the local store to end up
// stale relative to the homeserver after a restart (see IsAccountKeyMismatch).
func (s *Service) Close() error {
	if s == nil || s.helper == nil {
		return nil
	}
	return s.helper.Close()
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
