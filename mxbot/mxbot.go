package mxbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"time"
)

type Bot interface {
	Name() string

	EventHandlers() []EventHandler
	AddEventHandler(handler EventHandler)

	AddCommand(command *Command, filters ...Filter)

	Client() *mautrix.Client

	StartListening(ctx context.Context) error
	StopListening(ctx context.Context) error
	GetSyncer() mautrix.Syncer
	Filters() []Filter

	SendMessage(ctx context.Context, msg *Message) error
	JoinRoom(ctx context.Context, roomID id.RoomID) error
	Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error)
	StartTyping(ctx context.Context, roomID id.RoomID) error
	StopTyping(ctx context.Context, roomID id.RoomID) error
}

// BotCredentials - credentials of the bot for Matrix
// should be provided by the user
// (username, password, homeserverURL)
type BotCredentials struct {
	Username      string
	Password      string
	HomeServerURL string
}

var (
	defaultSyncerRetryTime = 5 * time.Second
	defaultTypingTimeout   = 10 * time.Second
)

// NewDefaultBot - Bot constructor
// botName - name of the bot (should be unique for engine)
// botCredentials - matrix credentials of the bot
func NewDefaultBot(botName string, botCredentials *BotCredentials) (Bot, error) {
	client, err := mautrix.NewClient(botCredentials.HomeServerURL, "", "")
	if err != nil {
		return nil, err
	}

	defaultFilters := []Filter{
		FilterNotMe(client),          // ignore messages from the bot itself
		FilterAfterStart(time.Now()), // ignore messages that were sent before start time
	}

	bot := &DefaultBot{
		matrixClient:  client,
		name:          botName,
		eventHandlers: make([]EventHandler, 0),
		filters:       defaultFilters,
		credentials:   botCredentials,
		logger:        slog.Default().With("bot", botName),

		syncerTimeRetry: defaultSyncerRetryTime,
		typingTimeout:   defaultTypingTimeout,
	}

	return bot, nil
}

var _ Bot = (*DefaultBot)(nil)

// DefaultBot - Bot implementation
type DefaultBot struct {
	name          string
	eventHandlers []EventHandler

	filters []Filter

	matrixClient *mautrix.Client

	credentials *BotCredentials

	logger *slog.Logger

	syncerTimeRetry time.Duration
	typingTimeout   time.Duration
}

func (b *DefaultBot) Name() string {
	return b.name
}

func (b *DefaultBot) EventHandlers() []EventHandler {
	return b.eventHandlers
}

func (b *DefaultBot) AddEventHandler(handler EventHandler) {
	b.eventHandlers = append(b.eventHandlers, handler)
}

func (b *DefaultBot) AddCommand(command *Command, filters ...Filter) {
	b.eventHandlers = append(
		b.eventHandlers,
		NewMessageHandler(
			command.Handler,
			append(filters, FilterCommand(command))...,
		),
	)
}

func (b *DefaultBot) Client() *mautrix.Client {
	return b.matrixClient
}

// Download - downloads the content of the mxc URL
func (b *DefaultBot) Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error) {
	return b.matrixClient.DownloadBytes(ctx, mxcURL)
}

func (b *DefaultBot) StartListening(ctx context.Context) error {

	if err := b.prepareBot(ctx); err != nil {
		return err
	}

	if err := b.startSyncer(ctx); err != nil {
		return err
	}

	return nil
}

func (b *DefaultBot) StopListening(ctx context.Context) error {

	_, err := b.matrixClient.Logout(ctx)

	return err
}

func (b *DefaultBot) Filters() []Filter {
	return b.filters
}

func (b *DefaultBot) GetSyncer() mautrix.Syncer {
	return b.matrixClient.Syncer
}

func (b *DefaultBot) JoinRoom(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.JoinRoomByID(ctx, roomID)

	return err
}

func (b *DefaultBot) SendMessage(ctx context.Context, msg *Message) error {
	if msg == nil {
		return ErrNilMessage
	}

	if msg.Text != "" {
		if err := b.sendTextMessage(msg.RoomID, msg.Text); err != nil {
			b.logger.Error("Failed to send message", "error", err)

			return errors.Join(err, ErrSendMessage)
		}
	}

	return nil
}

func (b *DefaultBot) sendTextMessage(roomID id.RoomID, text string) error {
	_, err := b.matrixClient.SendMessageEvent(context.Background(), roomID, event.EventMessage, event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          text,
		Format:        event.FormatHTML,
		FormattedBody: fmt.Sprintf("<p>%s</p>", text),
	})

	return err
}

func (b *DefaultBot) startSyncer(ctx context.Context) error {

	syncer := b.matrixClient.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		if !b.checkFilters(evt) {
			return
		}

		if err := b.StartTyping(ctx, evt.RoomID); err != nil {
			slog.Error("failed to start typing", "err", err)
		}

		defer func() {
			if err := b.StopTyping(ctx, evt.RoomID); err != nil {
				slog.Error("failed to stop typing", "err", err)
			}
		}()

		eventContext := NewDefaultCtx(ctx, evt, b)

		for _, handler := range b.eventHandlers {
			err := handler.Handle(eventContext)
			if err != nil {
				return
			}
		}
	})

	// Start goroutine for syncing the events from the homeserver
	go func(ctx context.Context) {
		b.logger.Info("syncer started")
		for {
			select {
			case <-ctx.Done():
				b.logger.Info("syncer stopped")
				return

			default:
				err := b.matrixClient.Sync()
				if err != nil {
					slog.Error("sync error", "error", err)
					time.Sleep(b.syncerTimeRetry) // Wait before retrying
				}
			}
		}
	}(ctx)

	return nil
}

// prepareBot - Authenticates the bot with the homeserver
// and register the bot if it is not registered
// also refreshes the access token if it is expired
func (b *DefaultBot) prepareBot(ctx context.Context) error {

	if err := b.authBot(ctx); err != nil {
		if err := b.registerBot(ctx); err != nil {
			return err
		}
	}

	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(3 * time.Minute)

		for range ticker.C {
			if err := b.authBot(ctx); err != nil {
				slog.Error("failed to auth bot", "error", err)
			}
		}
	}()

	return nil
}

// authBot - Authenticates the bot with the homeserver
func (b *DefaultBot) authBot(ctx context.Context) error {

	resp, err := b.matrixClient.Login(ctx, &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: b.credentials.Username,
		},
		Password: b.credentials.Password,
	})

	if err != nil {
		return err
	}

	b.matrixClient.UserID = resp.UserID
	b.matrixClient.AccessToken = resp.AccessToken

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

// checkFilters - Checks if the event should be processed
func (b *DefaultBot) checkFilters(evt *event.Event) bool {
	for _, filter := range b.filters {
		if !filter(evt) {
			return false
		}
	}
	return true
}

// StartTyping - Starts typing on the room
// it will be stopped after timeout or when user stops typing
func (b *DefaultBot) StartTyping(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.UserTyping(ctx, roomID, true, b.typingTimeout)
	return err
}

// StopTyping - Stops typing on the room
func (b *DefaultBot) StopTyping(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.UserTyping(ctx, roomID, false, b.typingTimeout)
	return err
}
