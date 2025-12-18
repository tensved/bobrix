package mxbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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

	GetJoinedRoomsList(ctx context.Context) ([]id.RoomID, error)
	GetMessagesFromRoomByNumber(ctx context.Context, roomID id.RoomID, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error)
	GetMessagesFromRoomByDuration(ctx context.Context, roomID id.RoomID, duration time.Duration, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error)

	DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error)
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

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	bot := &DefaultBot{
		matrixClient:  client,
		name:          botName,
		eventHandlers: make([]EventHandler, 0),
		credentials:   botCredentials,
		logger:        &logger,

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
	machine      *crypto.OlmMachine

	botStatus event.Presence

	credentials *BotCredentials

	logger *zerolog.Logger

	syncerTimeRetry time.Duration
	typingTimeout   time.Duration

	cancelFunc context.CancelFunc
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
	syncCtx, cancel := context.WithCancel(ctx)
	b.cancelFunc = cancel

	if err := b.prepareBot(ctx); err != nil {
		return err
	}

	if err := b.startSyncer(syncCtx); err != nil {
		return err
	}

	return nil
}

func (b *DefaultBot) StopListening(ctx context.Context) error {
	if b.cancelFunc != nil {
		b.cancelFunc() // stops goroutine с Sync()
	}

	b.matrixClient.StopSync()

	b.logger.Info().Msg("stop sync and logout")

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

	// Checking if the room is encrypted
	var encryptionEvent event.EncryptionEventContent
	err := b.matrixClient.StateEvent(ctx, roomID, event.StateEncryption, "", &encryptionEvent)
	if err != nil {
		// If get 404, it means the room is not encrypted.
		if err.Error() == "M_NOT_FOUND (HTTP 404): Event not found." {
			// Send a not encrypted message
			_, err = b.matrixClient.SendMessageEvent(ctx, roomID, event.EventMessage, msg.AsJSON())
			if err != nil {
				return fmt.Errorf("%w: %w", ErrSendMessage, err)
			}
			return nil
		}
		slog.Error("failed to get room encryption state", "error", err)
		return fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	// Let's check if we have a session for this room
	session, err := b.machine.CryptoStore.GetOutboundGroupSession(ctx, roomID)
	if err != nil || session == nil {
		// If there is no session, create a new one
		// Get a list of room participants
		members, err := b.matrixClient.Members(ctx, roomID)
		if err != nil {
			slog.Error("failed to get room members", "error", err)
			return fmt.Errorf("%w: %w", ErrSendMessage, err)
		}

		// Collecting a list of user IDs
		userIDs := make([]id.UserID, 0, len(members.Chunk))
		for _, member := range members.Chunk {
			if member.StateKey != nil {
				userIDs = append(userIDs, id.UserID(*member.StateKey))
			}
		}

		err = b.machine.ShareGroupSession(ctx, roomID, userIDs)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrSendMessage, err)
		}
	}

	// Determine event type based on message type
	eventType := event.EventMessage
	if msg.Type().IsMedia() {
		switch msg.Type() {
		case event.MsgImage:
			eventType = event.EventMessage
		case event.MsgVideo:
			eventType = event.EventMessage
		case event.MsgAudio:
			eventType = event.EventMessage
		case event.MsgFile:
			eventType = event.EventMessage
		}
	}

	encryptedContent, err := b.machine.EncryptMegolmEvent(ctx, roomID, eventType, msg.AsJSON())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendMessage, err)
	}

	// If encryption was successful, send the encrypted message
	_, err = b.matrixClient.SendMessageEvent(ctx, roomID, event.EventEncrypted, encryptedContent)
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
		b.logger.Warn().Err(err).Msg("failed to start typing")
	} else {
		defer cancelTyping()
	}

	eventContext, err := NewDefaultCtx(ctx, evt, b)

	defer eventContext.SetHandled()

	if err != nil {
		b.logger.Error().Err(err).Msg("failed to create event context")
		return
	}

	for _, handler := range b.eventHandlers {
		err := handler.Handle(eventContext)
		if err != nil {
			return
		}
	}
}

func (b *DefaultBot) DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error) {
	return b.machine.DecryptMegolmEvent(ctx, evt)
}

func (b *DefaultBot) startSyncer(ctx context.Context) error {
	syncer := b.matrixClient.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEvent(func(eventCtx context.Context, evt *event.Event) {
		go b.eventHandler(eventCtx, evt)
	})

	// Encrypted Message Handler
	syncer.OnEventType(event.EventEncrypted, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Msg("received encrypted message")

		_, err := b.machine.DecryptMegolmEvent(ctx, evt)
		if err != nil {
			b.logger.Error().
				Err(err).
				Str("room_id", evt.RoomID.String()).
				Str("sender", string(evt.Sender)).
				Msg("failed to decrypt message")

			if err.Error() == "no session with given ID found" {
				// Request the key from the sender
				err = b.machine.SendRoomKeyRequest(ctx, evt.RoomID, evt.Content.AsEncrypted().SenderKey, evt.Content.AsEncrypted().SessionID, "m.megolm.v1", map[id.UserID][]id.DeviceID{
					evt.Sender: {id.DeviceID(evt.Content.AsEncrypted().DeviceID)},
				})
				if err != nil {
					b.logger.Error().Err(err).Msg("failed to request room key")
					return
				}
				b.logger.Info().Msg("sent room key request")
			}
			return
		}

		b.logger.Info().Msg("message decrypted successfully")
	})

	// Room Key Receiver Handler
	syncer.OnEventType(event.ToDeviceRoomKey, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Interface("content", evt.Content.Raw).Msg("received room key event")
		content := evt.Content.AsRoomKey()
		if content == nil {
			b.logger.Error().Msg("invalid room key format")
			return
		}

		b.machine.HandleToDeviceEvent(ctx, evt)
	})

	// Key Request Handler
	syncer.OnEventType(event.ToDeviceRoomKeyRequest, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Msg("received room key request")
		content := evt.Content.AsRoomKeyRequest()
		if content == nil {
			b.logger.Error().Msg("invalid room key request format")
			return
		}

		b.logger.Info().Str("sender", string(evt.Sender)).Msg("room key request from")
		b.machine.HandleToDeviceEvent(ctx, evt)
	})

	// Handler for receiving forwarded room keys
	syncer.OnEventType(event.ToDeviceForwardedRoomKey, func(ctx context.Context, evt *event.Event) {
		b.logger.Info().Msg("received forwarded room key")
		content := evt.Content.AsForwardedRoomKey()
		if content == nil {
			b.logger.Error().Msg("invalid forwarded room key format")
			return
		}

		b.machine.HandleToDeviceEvent(ctx, evt)
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				b.logger.Info().Msg("syncer stopped by context")
				return

			default:
				b.logger.Info().Msg("start sync")
				err := b.matrixClient.SyncWithContext(ctx)

				if ctx.Err() != nil {
					return
				}

				if err != nil {
					httpErr, ok := err.(mautrix.HTTPError)

					if ok && httpErr.RespError.StatusCode == 401 {
						b.logger.Warn().Msg("token expired, reauth needed")

						if err := b.authBot(ctx); err != nil {
							b.logger.Error().Err(err).Msg("failed to reauth bot")
							time.Sleep(b.syncerTimeRetry)
							continue
						}

						b.logger.Info().Msg("reauth success, restarting sync")
						continue
					}

					b.logger.Error().Err(err).Msg("failed to sync")
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
	if err := b.authorizeBot(ctx); err != nil {
		return err
	}

	if b.displayName != "" {
		if err := b.matrixClient.SetDisplayName(ctx, b.displayName); err != nil {
			return err
		}
	}

	// Initialize crypto after successful login
	if err := b.initCrypto(ctx); err != nil {
		return fmt.Errorf("failed to init crypto: %w", err)
	}

	return nil
}

func (b *DefaultBot) authorizeBot(ctx context.Context) error {
	if err := b.authBot(ctx); err != nil {
		if err := b.registerBot(ctx); err != nil {
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
					b.logger.Error().Err(err).Msg("failed to stop typing")
				}

			case <-c:
				typing.remove(roomID)

				if err := b.StopTyping(ctx, roomID); err != nil {
					b.logger.Error().Err(err).Msg("failed to stop typing")
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

// GetThreadStory - gets the thread story
func (b *DefaultBot) GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*MessagesThread, error) {

	msgs, err := b.matrixClient.Messages(
		ctx,
		roomID,
		"",
		"",
		mautrix.DirectionBackward,
		nil,
		b.credentials.ThreadLimit,
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
			slog.Debug("rel type not thread", "event_id", evt.ID, "rel_event_id", rel.EventID, "rel_type", rel.Type)
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

func (b *DefaultBot) GetJoinedRoomsList(ctx context.Context) ([]id.RoomID, error) {
	rooms, err := b.matrixClient.JoinedRooms(ctx)
	return rooms.JoinedRooms, err
}

func (b *DefaultBot) GetMessagesFromRoomByNumber(ctx context.Context, roomID id.RoomID, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error) {
	resp, err := b.matrixClient.Messages(ctx, roomID, "", "", 'b', filter, numMessages)
	if err != nil {
		return []*event.Event{}, err
	}

	messages := make([]*event.Event, 0, len(resp.Chunk))
	for i := len(resp.Chunk) - 1; i >= 0; i-- { // evts ascending order by the time
		messages = append(messages, resp.Chunk[i])

	}
	return messages, nil
}

func (b *DefaultBot) GetMessagesFromRoomByDuration(ctx context.Context, roomID id.RoomID, duration time.Duration, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error) {
	timeAgo := time.Now().Add(-1 * duration).UnixMilli()

	syncResp, err := b.matrixClient.SyncRequest(ctx, 0, "", "", false, "")
	if err != nil {
		return nil, err
	}
	from := syncResp.NextBatch

	var allMessages []*event.Event

	for {
		msgsResp, err := b.matrixClient.Messages(ctx, roomID, from, "", 'b', filter, numMessages)
		if err != nil {
			return nil, err
		}

		stop := false
		for _, evt := range msgsResp.Chunk {
			if evt.Timestamp <= timeAgo {
				stop = true
				break
			}
			allMessages = append(allMessages, evt)
		}

		if stop || msgsResp.End == "" {
			break
		}

		from = msgsResp.End
	}

	slices.Reverse(allMessages) // evts ascending order by the time

	return allMessages, nil
}