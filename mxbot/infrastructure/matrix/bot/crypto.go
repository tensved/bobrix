package bot // nok

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
)

var _ domainbot.Crypto = (*MatrixCrypto)(nil)

type MatrixCrypto struct {
	olmMachine OlmMachine
}

func (c *MatrixCrypto) Decrypt(
	ctx context.Context,
	evt *event.Event,
) (*event.Event, error) {

	if !c.IsEncrypted(evt) {
		return evt, nil
	}

	decrypted, err := c.olmMachine.DecryptEvent(ctx, evt)
	if err != nil {
		return nil, fmt.Errorf("decrypt event %s: %w", evt.ID, err)
	}

	return decrypted, nil
}

func (c *MatrixCrypto) IsEncrypted(evt *event.Event) bool {
	return evt.Type == event.EventEncrypted
}

// ----------------------------

func (b *DefaultBot) initCrypto(ctx context.Context) error {
	// Check that the client is authorized
	if b.matrixClient.UserID == "" || b.matrixClient.AccessToken == "" {
		return fmt.Errorf("client is not logged in")
	}

	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	storeDir := filepath.Join(currentDir, ".bin", "crypto", fmt.Sprintf("crypto-store-%s.db", b.name))

	// Create a directory to store cryptographic data
	cryptoDir := filepath.Join(currentDir, ".bin", "crypto")
	if err := os.MkdirAll(cryptoDir, 0755); err != nil {
		return fmt.Errorf("failed to create crypto directory: %w", err)
	}

	// Create a crypto helper with automatic session management
	cryptoHelper, err := cryptohelper.NewCryptoHelper(b.matrixClient, b.credentials.PickleKey, storeDir)
	if err != nil {
		return fmt.Errorf("failed to create crypto helper: %w", err)
	}

	// Initialize the crypto helper
	err = cryptoHelper.Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to init crypto helper: %w", err)
	}

	// We get a machine for encryption/decryption
	b.machine = cryptoHelper.Machine()

	// Loading the machine context
	err = b.machine.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load olm machine: %w", err)
	}

	identity := b.machine.OwnIdentity()
	if identity == nil {
		return fmt.Errorf("failed to get own identity")
	}

	b.logger.Info().Interface("identity", identity).Msg("crypto initialized")

	return nil
}

func (b *DefaultBot) DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error) {
	return b.machine.DecryptMegolmEvent(ctx, evt)
}
