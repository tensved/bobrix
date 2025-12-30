package bot

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/tensved/bobrix/mxbot"
	"maunium.net/go/mautrix"
)

func NewDefault(
	name string,
	creds *BotCredentials,
	opts ...mxbot.BotOption,
) (*DefaultBot, error) {

	client, err := mautrix.NewClient(creds.HomeServerURL, "", "")
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	b := &DefaultBot{
		name:   name,
		client: client,
		logger: &logger,
	}

	b.auth = newAuthService(client, creds, name)
	b.crypto = newCryptoService(client, creds, name, &logger)
	b.typing = newTypingService(client, time.Second*30)
	b.threads = newThreadService(client, creds.ThreadLimit)
	b.history = newHistoryService(client)

	b.syncer = newSyncService(
		client,
		b.crypto,
		b.dispatchEvent,
		b.reauth,
		time.Second*5,
		&logger,
	)

	return b, nil
}
