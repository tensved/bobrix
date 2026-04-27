package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	infracfg "github.com/tensved/bobrix/mxbot/infrastructure/matrix/config"
	utils "github.com/tensved/bobrix/mxbot/infrastructure/utils"
)

var _ dombot.BotAuth = (*Service)(nil)

type Service struct {
	client *mautrix.Client
	creds  *infracfg.BotCredentials
	name   string
}

func New(client *mautrix.Client, creds *infracfg.BotCredentials, name string) (*Service, error) {
	if name == "" {
		return nil, errors.New("bot name shouldnt be an empty string")
	}
	return &Service{
		client: client,
		creds:  creds,
		name:   name,
	}, nil
}

func (a *Service) Authorize(ctx context.Context) error {
	if err := a.authBot(ctx); err == nil {
		return nil
	}

	// login failed - trying to create/update user
	if err := a.registerBot(ctx); err != nil {
		return err
	}

	return a.authBot(ctx)
}

// authBot - Authenticates the bot with the homeserver
func (a *Service) authBot(ctx context.Context) error {
	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	safeUser := utils.SafeFilePart(a.creds.Username)

	// Check if a file with a saved device ID exists
	deviceIDFile := filepath.Join(currentDir, ".bin", "crypto", fmt.Sprintf("device-id-%s.txt", safeUser))
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
		return fmt.Errorf("error login: %w", err)
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

// оставляем буквы/цифры/._-@, остальное заменяем на _
var reUnsafe = regexp.MustCompile(`[^a-zA-Z0-9._\-@]+`)

func safeFilePart(s string) string {
	s = strings.TrimSpace(s)
	// на всякий случай убираем любые разделители пути текущей ОС
	s = strings.ReplaceAll(s, string(filepath.Separator), "_")
	// и второй разделитель (актуально для Windows, где могут встретиться оба)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")

	s = reUnsafe.ReplaceAllString(s, "_")
	if s == "" {
		s = "unknown"
	}
	return s
}
