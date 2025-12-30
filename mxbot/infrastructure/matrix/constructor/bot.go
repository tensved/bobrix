package constructor

import (
	"fmt"

	// "github.com/docker/docker/daemon/events"
	// "github.com/docker/docker/daemon/events"
	"github.com/tensved/bobrix/mxbot/domain/bot"

	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/client"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/events"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/info"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/messaging"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/rooms"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/threads"
	"github.com/tensved/bobrix/mxbot/infrastructure/matrix/typing"
)

type MatrixBot struct {
	bot.BotInfo
	bot.BotMessaging
	bot.BotThreads
	bot.BotCrypto
	bot.BotClient
	bot.EventLoader
	bot.BotRoomActions
	bot.BotTyping
}

func NewMatrixBot(cfg Config) (*MatrixBot, error) {
	client, err := client.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("dd")
	}

	rooms := rooms.New(client)
	messaging := messaging.New(client)
	typing := typing.New(client, cfg.typingTimeout, cfg.logger)
	info := info.New(client, cfg.Name)

	crypto := crypto.New(client, cfg.Crypto)
	events := events.New(client)
	threads := threads.New(client, events)

	return &MatrixBot{
		BotInfo:        info,
		BotMessaging:   messaging,
		BotCrypto:      crypto,
		BotClient:      client,
		EventLoader:    events,
		BotThreads:     threads,
		BotRoomActions: rooms,
		BotTyping:      typing,
	}, nil
}

// BotCredentials - credentials of the bot for Matrix
// should be provided by the user
// (username, password, homeserverURL)
type BotCredentials struct {
	Username      string
	Password      string
	HomeServerURL string
	PickleKey     []byte
	ThreadLimit   int
	AuthMode      AuthMode
	ASToken       string
}
