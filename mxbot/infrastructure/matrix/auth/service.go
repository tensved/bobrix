package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
)

var _ dombot.BotAuth = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	creds  *infracfg.BotCredentials
	name   string
}

func New(client *mautrix.Client, creds *infracfg.BotCredentials, name string) *Service {
	return &Service{
		client: client,
		creds:  creds,
		name:   name,
	}
}

func (a *Service) Authorize(ctx context.Context) error {
	if err := a.authBot(ctx); err != nil {
		if err := a.registerBot(ctx); err != nil {
			return err
		}
	}
	return nil
}

// authBot - Authenticates the bot with the homeserver
func (a *Service) authBot(ctx context.Context) error {
	// Получаем текущую директорию
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if a file with a saved device ID exists
	deviceIDFile := filepath.Join(currentDir, ".bin", "crypto", fmt.Sprintf("device-id-%s.txt", a.creds.Username))
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
			User: a.creds.Username,
		},
		Password: a.creds.Password,
	}

	// If we have a saved device ID, we use it
	if deviceID != "" {
		loginReq.DeviceID = deviceID
	}

	resp, err := a.client.Login(ctx, loginReq)
	if err != nil {
		return err
	}

	a.client.UserID = resp.UserID
	a.client.AccessToken = resp.AccessToken
	a.client.DeviceID = resp.DeviceID

	// Save device ID to file
	cryptoDir := filepath.Join(currentDir, ".bin", "crypto")
	if err := os.MkdirAll(cryptoDir, 0755); err != nil {
		return fmt.Errorf("failed to create crypto directory: %w", err)
	}

	if err := os.WriteFile(deviceIDFile, []byte(resp.DeviceID), 0644); err != nil {
		return fmt.Errorf("failed to save device ID: %w", err)
	}

	// We check that the client is actually authorized
	whoami, err := a.client.Whoami(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify login: %w", err)
	}

	if whoami.UserID != resp.UserID {
		return fmt.Errorf("user ID mismatch: got %s, expected %s", whoami.UserID, resp.UserID)
	}

	return nil
}

// registerBot - Registers the bot with the homeserver
func (a *Service) registerBot(ctx context.Context) error {
	resp, err := a.client.RegisterDummy(ctx, &mautrix.ReqRegister{
		Username:     a.creds.Username,
		Password:     a.creds.Password,
		InhibitLogin: false,
		Auth:         nil,
		Type:         mautrix.AuthTypeDummy,
	})
	if err != nil {
		return err
	}

	a.client.UserID = resp.UserID
	a.client.AccessToken = resp.AccessToken
	a.client.DeviceID = resp.DeviceID

	return nil
}
