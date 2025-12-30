package bot // nok

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type AuthMode string

const (
	AuthModeLogin AuthMode = "login"
	AuthModeAS    AuthMode = "as"
)

type authService struct {
	client *mautrix.Client
	creds  *BotCredentials
	name   string
}

func newAuthService(c *mautrix.Client, creds *BotCredentials, name string) *authService {
	return &authService{client: c, creds: creds, name: name}
}

func (a *authService) AuthorizeBot(ctx context.Context) error {
	switch a.creds.AuthMode {
	case AuthModeAS:
		return a.authorizeAs(ctx)
	default:
		return a.authorizeBotViaLogin(ctx)
	}
}

func (a *authService) authorizeAs(ctx context.Context) error {
	userID := "@" + a.credentials.Username + ":" + a.matrixClient.HomeserverURL.Host

	if a.authMode == AuthModeAS {
		a.matrixClient.UserID = id.UserID(userID)
		a.matrixClient.AccessToken = a.asToken
		a.matrixClient.DeviceID = "AS_DEVICE"

		return nil
	}

	return a.authorizeViaLogin(ctx)
}

func (a *authService) authorizeBotViaLogin(ctx context.Context) error {
	if err := a.authBot(ctx); err != nil {
		if err := a.registerBot(ctx); err != nil {
			return err
		}
	}
	return nil
}

// authBot - Authenticates the bot with the homeserver
func (b *DefaultBot) authBot(ctx context.Context) error {
	// Получаем текущую директорию
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if a file with a saved device ID exists
	deviceIDFile := filepath.Join(currentDir, ".bin", "crypto", fmt.Sprintf("device-id-%s.txt", b.name))
	var deviceID id.DeviceID

	if _, err := os.Stat(deviceIDFile); err == nil {
		// If the file exists, read the device ID
		data, err := os.ReadFile(deviceIDFile)
		if err != nil {
			return fmt.Errorf("failed to read device ID file: %w", err)
		}

		deviceID = id.DeviceID(string(data))
	}

	loginReq := &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: b.credentials.Username,
		},
		Password: b.credentials.Password,
	}

	// If we have a saved device ID, we use it
	if deviceID != "" {
		loginReq.DeviceID = deviceID
	}

	resp, err := b.matrixClient.Login(ctx, loginReq)
	if err != nil {
		return err
	}

	b.matrixClient.UserID = resp.UserID
	b.matrixClient.AccessToken = resp.AccessToken
	b.matrixClient.DeviceID = resp.DeviceID

	// Save device ID to file
	cryptoDir := filepath.Join(currentDir, ".bin", "crypto")
	if err := os.MkdirAll(cryptoDir, 0755); err != nil {
		return fmt.Errorf("failed to create crypto directory: %w", err)
	}

	if err := os.WriteFile(deviceIDFile, []byte(resp.DeviceID), 0644); err != nil {
		return fmt.Errorf("failed to save device ID: %w", err)
	}

	// We check that the client is actually authorized
	whoami, err := b.matrixClient.Whoami(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify login: %w", err)
	}

	if whoami.UserID != resp.UserID {
		return fmt.Errorf("user ID mismatch: got %s, expected %s", whoami.UserID, resp.UserID)
	}

	return nil
}

// registerBot - Registers the bot with the homeserver
func (b *DefaultBot) registerBot(ctx context.Context) error {
	resp, err := b.matrixClient.RegisterDummy(ctx, &mautrix.ReqRegister{
		Username:     b.credentials.Username,
		Password:     b.credentials.Password,
		InhibitLogin: false,
		Auth:         nil,
		Type:         mautrix.AuthTypeDummy,
	})
	if err != nil {
		return err
	}

	b.matrixClient.UserID = resp.UserID
	b.matrixClient.AccessToken = resp.AccessToken

	return nil
}
