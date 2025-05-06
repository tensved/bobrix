package mxbot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tensved/bobrix/mxbot/messages"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"slices"
	"sync"
	"time"
)

type Bot interface {
	Name() string
	FullName() string // get full name with servername (e.g. @username:servername)
	UserID() id.UserID

	EventHandlers() []EventHandler
	AddEventHandler(handler EventHandler)

	AddCommand(command *Command, filters ...Filter)

	Client() *mautrix.Client

	StartListening(ctx context.Context) error
	StopListening(ctx context.Context) error
	GetSyncer() mautrix.Syncer
	Filters() []Filter
	AddFilter(filter Filter)

	SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error
	JoinRoom(ctx context.Context, roomID id.RoomID) error
	Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error)
	StartTyping(ctx context.Context, roomID id.RoomID) error
	StopTyping(ctx context.Context, roomID id.RoomID) error

	Ping(ctx context.Context) error

	IsThreadEnabled() bool

	GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*MessagesThread, error)
	GetThreadByEvent(ctx context.Context, evt *event.Event) (*MessagesThread, error)

	SetIdleStatus()
	SetOnlineStatus()
	SetOfflineStatus()
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
	defaultTypingTimeout   = 30 * time.Second
)

type BotOptions func(*DefaultBot) // Bot options. Used to configure the bot

// WithSyncerRetryTime - time to wait before retrying a failed sync
func WithSyncerRetryTime(time time.Duration) BotOptions {
	return func(bot *DefaultBot) {
		bot.syncerTimeRetry = time
	}
}

// WithTypingTimeout - time to wait before sending a typing event
func WithTypingTimeout(time time.Duration) BotOptions {
	return func(bot *DefaultBot) {
		bot.typingTimeout = time
	}
}

func WithDisplayName(name string) BotOptions {
	return func(bot *DefaultBot) {
		bot.displayName = name
	}
}

// NewDefaultBot - Bot constructor
// botName - name of the bot (should be unique for engine)
// botCredentials - matrix credentials of the bot
func NewDefaultBot(botName string, botCredentials *BotCredentials, opts ...BotOptions) (Bot, error) {
	client, err := mautrix.NewClient(botCredentials.HomeServerURL, "", "")
	if err != nil {
		return nil, err
	}

	bot := &DefaultBot{
		matrixClient:  client,
		name:          botName,
		eventHandlers: make([]EventHandler, 0),
		credentials:   botCredentials,
		logger:        slog.Default().With("bot", botName),

		syncerTimeRetry: defaultSyncerRetryTime,
		typingTimeout:   defaultTypingTimeout,
	}

	defaultFilters := []Filter{
		FilterNotMe(client), // ignore messages from the bot itself
		FilterAfterStart(
			bot,
			FilterAfterStartOptions{
				StartTime:      time.Now(),
				ProcessInvites: true,
			}), // ignore messages that were sent before start time
	}

	bot.filters = defaultFilters

	for _, opt := range opts {
		opt(bot)
	}

	return bot, nil
}

var _ Bot = (*DefaultBot)(nil)

// DefaultBot - Bot implementation
type DefaultBot struct {
	name          string
	displayName   string
	eventHandlers []EventHandler

	filters []Filter

	isThreadEnabled bool // thread support

	matrixClient *mautrix.Client

	botStatus event.Presence

	credentials *BotCredentials

	logger *slog.Logger

	syncerTimeRetry time.Duration
	typingTimeout   time.Duration
}

func (b *DefaultBot) Name() string {
	return b.name
}

func (b *DefaultBot) FullName() string {
	return b.Client().UserID.String()
}

func (b *DefaultBot) UserID() id.UserID {
	return b.Client().UserID
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

	b.matrixClient.StopSync()

	b.logger.Info("stop sync and logout")

	_, err := b.matrixClient.Logout(ctx)

	return err
}

func (b *DefaultBot) Filters() []Filter {
	return b.filters
}

func (b *DefaultBot) AddFilter(filter Filter) {
	b.filters = append(b.filters, filter)
}

func (b *DefaultBot) GetSyncer() mautrix.Syncer {
	return b.matrixClient.Syncer
}

func (b *DefaultBot) JoinRoom(ctx context.Context, roomID id.RoomID) error {
	_, err := b.matrixClient.JoinRoomByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("%w roomID=%v: %w", ErrJoinToRoom, roomID, err)
	}

	return nil
}

func (b *DefaultBot) SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error {
	if msg == nil {
		return ErrNilMessage
	}

	if msg.Type().IsMedia() {
		uploadResponse, err := b.matrixClient.UploadMedia(ctx, msg.AsReqUpload())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrUploadMedia, err)
		}
		msg.SetContentURI(uploadResponse.ContentURI)
	}

	_, err := b.matrixClient.SendMessageEvent(ctx, roomID, event.EventMessage, msg.AsJSON())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	return nil
}

func (b *DefaultBot) eventHandler(ctx context.Context, evt *event.Event) {

	if !b.checkFilters(evt) {
		return
	}

	cancelTyping, err := b.LoopTyping(ctx, evt.RoomID)

	if err != nil {
		b.logger.Warn("failed to start typing", "err", err)
	} else {
		defer cancelTyping()
	}

	eventContext, err := NewDefaultCtx(ctx, evt, b)

	defer eventContext.SetHandled()

	if err != nil {
		b.logger.Error("failed to create event context", "err", err)
		return
	}

	for _, handler := range b.eventHandlers {
		err := handler.Handle(eventContext)
		if err != nil {
			return
		}
	}
}

func (b *DefaultBot) startSyncer(ctx context.Context) error {
	syncer := b.matrixClient.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEvent(func(eventCtx context.Context, evt *event.Event) {
		go b.eventHandler(eventCtx, evt)
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				b.logger.Info("syncer stopped by context")
				return
			default:
				b.logger.Info("start sync")
				if err := b.matrixClient.Sync(); err != nil {
					b.logger.Error("failed to sync", "err", err)
					time.Sleep(b.syncerTimeRetry)
				}
			}
		}
	}()

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

	if b.displayName != "" {
		if err := b.matrixClient.SetDisplayName(ctx, b.displayName); err != nil {
			return err
		}
	}

	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(3 * time.Minute)

		for range ticker.C {
			if err := b.authBot(ctx); err != nil {
				b.logger.Error("failed to auth bot", "error", err)
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

// typingData - struct for store information about typing by bots in rooms
// it is used to avoid problem with cancelling typingEvent when bot have multiple requests at the same time
type typingData struct {
	data map[string]int
	mx   *sync.RWMutex
}

func (d *typingData) add(roomID id.RoomID) {
	d.mx.Lock()
	d.data[roomID.String()]++
	d.mx.Unlock()
}

func (d *typingData) remove(roomID id.RoomID) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.data[roomID.String()]--

	if d.data[roomID.String()] == 0 {
		delete(d.data, roomID.String())
	}
}

var typing = typingData{
	data: make(map[string]int),
	mx:   &sync.RWMutex{},
}

// LoopTyping - Starts and stops typing on the room. Typing is sent every typingTimeout
// it will be stopped if the context is cancelled or if an error occurs
// it returns a function that can be used to stop the typing
func (b *DefaultBot) LoopTyping(ctx context.Context, roomID id.RoomID) (cancelTyping func(), err error) {

	ticker := time.NewTicker(b.typingTimeout)

	typing.add(roomID)

	err = b.StartTyping(ctx, roomID)
	if err != nil {
		return nil, err
	}

	cancel := make(chan struct{})

	go func(c <-chan struct{}) {
		for {
			select {
			case <-ticker.C:
				if err := b.StartTyping(ctx, roomID); err != nil {
					b.logger.Error("failed to stop typing", "err", err)
				}

			case <-c:
				typing.remove(roomID)

				if err := b.StopTyping(ctx, roomID); err != nil {
					b.logger.Error("failed to stop typing", "err", err)
				}

				return
			}
		}
	}(cancel)

	return func() {
		cancel <- struct{}{}
		ticker.Stop()

		close(cancel)
	}, nil

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

// Ping - Checks if the bot is online
// It will return error if the bot is offline
func (b *DefaultBot) Ping(ctx context.Context) error {
	_, err := b.matrixClient.GetOwnDisplayName(ctx)
	return err
}

func (b *DefaultBot) IsThreadEnabled() bool {
	return b.isThreadEnabled
}

func (b *DefaultBot) GetThreadByEvent(ctx context.Context, evt *event.Event) (*MessagesThread, error) {

	if evt == nil {
		return nil, ErrNilEvent
	}

	rel := evt.Content.AsMessage().RelatesTo

	if rel == nil || rel.Type != event.RelThread {
		return nil, nil
	}

	roomID := evt.RoomID
	parentEventID := rel.EventID

	return b.GetThread(ctx, roomID, parentEventID)
}

func (b *DefaultBot) SetStatus(status event.Presence) {
	b.botStatus = status
}

func (b *DefaultBot) SetIdleStatus() {
	b.SetStatus(event.PresenceUnavailable)
}

func (b *DefaultBot) SetOnlineStatus() {
	b.SetStatus(event.PresenceOnline)
}

func (b *DefaultBot) SetOfflineStatus() {
	b.SetStatus(event.PresenceOffline)
}

const threadLimit = 120

// GetThreadStory - gets the thread story
func (b *DefaultBot) GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*MessagesThread, error) {

	msgs, err := b.matrixClient.Messages(
		ctx,
		roomID,
		"",
		"",
		mautrix.DirectionBackward,
		nil,
		threadLimit,
	)
	if err != nil {
		slog.Error("error get messages", "error", err)
		return nil, err
	}

	data := make([]*event.Event, 0)

	for _, evt := range msgs.Chunk {

		msg, ok := GetFixedMessage(evt)
		if !ok {
			continue
		}

		evt.Content.Parsed = msg

		if evt.ID == parentEventID {

			data = append(data, evt)

			rawContent := evt.Content.Raw

			if answerEventID, ok := rawContent[AnswerToCustomField]; ok {

				evtID := id.EventID(answerEventID.(string))

				answerEvent, err := b.matrixClient.GetEvent(ctx, roomID, evtID)
				if err != nil {
					slog.Error("error get answer event", "error", err)
					break
				}

				answerMsg, ok := GetFixedMessage(answerEvent)
				if !ok {
					continue
				}

				answerEvent.Content.Parsed = answerMsg

				data = append(data, answerEvent)
			}
			break
		}

		rel := msg.RelatesTo

		if rel == nil {
			slog.Info("rel type nil", "event_id", evt.ID)
			continue
		}

		if rel.Type != event.RelThread || rel.EventID != parentEventID {
			slog.Info("rel type not thread", "event_id", evt.ID, "rel_event_id", rel.EventID, "rel_type", rel.Type)
			continue
		}

		data = append(data, evt)
	}

	slog.Info("data size", "size", len(data))

	slices.Reverse(data)

	thread := &MessagesThread{
		Messages: data,
		ParentID: parentEventID,
		RoomID:   roomID,
	}

	return thread, nil
}

func GetFixedMessage(evt *event.Event) (*event.MessageEventContent, bool) {
	veryRaw := evt.Content.VeryRaw

	if veryRaw == nil {
		return nil, false
	}

	var msg *event.MessageEventContent

	if err := json.Unmarshal(veryRaw, &msg); err != nil {
		slog.Error("error unmarshal message", "error", err)
		return nil, false
	}

	return msg, true
}
