package bot

import (
	"context"
	"time"

	// applfilters "github.com/tensved/bobrix/mxbot/application/filters"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"github.com/tensved/bobrix/mxbot/domain/threads"
	"github.com/tensved/bobrix/mxbot/messages"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// ----- BotAuth

func (b *DefaultBot) Authorize(ctx context.Context) error {
	return b.auth.Authorize(ctx)
}

// ----- BotInfo

func (b *DefaultBot) UserID() id.UserID {
	return b.info.UserID()
}

func (b *DefaultBot) FullName() string {
	return b.info.FullName()
}

func (b *DefaultBot) Name() string {
	return b.info.Name()
}

// ----- BotMessaging

func (b *DefaultBot) SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error {
	return b.messaging.SendMessage(ctx, roomID, msg)
}

// ----- BotThreads

func (b *DefaultBot) IsThreadEnabled() bool {
	return b.isThreadEnabled
}

func (b *DefaultBot) GetThreadByEvent(ctx context.Context, evt *event.Event) (*threads.MessagesThread, error) {
	return b.threads.GetThreadByEvent(ctx, evt)
}

func (b *DefaultBot) GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*threads.MessagesThread, error) {
	return b.threads.GetThread(ctx, roomID, parentEventID)
}

// ----- BotClient
func (b *DefaultBot) RawClient() any {
	return b.client.RawClient()
}

// ----- EventLoader

func (b *DefaultBot) GetEvent(ctx context.Context, roomID id.RoomID, eventID id.EventID) (*event.Event, error) {
	return b.eventLoader.GetEvent(ctx, roomID, eventID)
}

// ----- BotRoomActions

func (b *DefaultBot) JoinRoom(ctx context.Context, roomID id.RoomID) error {
	return b.rooms.JoinRoom(ctx, roomID)
}

func (b *DefaultBot) JoinedMembersCount(ctx context.Context, roomID id.RoomID) (int, error) {
	return b.rooms.JoinedMembersCount(ctx, roomID)
}

func (b *DefaultBot) GetJoinedRoomsList(ctx context.Context) ([]id.RoomID, error) {
	return b.rooms.GetJoinedRoomsList(ctx)
}

func (b *DefaultBot) GetMessagesFromRoomByNumber(ctx context.Context, roomID id.RoomID, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error) {
	return b.rooms.GetMessagesFromRoomByNumber(ctx, roomID, numMessages, filter)
}

func (b *DefaultBot) GetMessagesFromRoomByDuration(ctx context.Context, roomID id.RoomID, duration time.Duration, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error) {
	return b.rooms.GetMessagesFromRoomByDuration(ctx, roomID, duration, numMessages, filter)
}

// ----- BotTyping
func (b *DefaultBot) LoopTyping(ctx context.Context, roomID id.RoomID) (cancelTyping func(), err error) {
	return b.typing.LoopTyping(ctx, roomID)
}

func (b *DefaultBot) StartTyping(ctx context.Context, roomID id.RoomID) error {
	return b.typing.StartTyping(ctx, roomID)
}

func (b *DefaultBot) StopTyping(ctx context.Context, roomID id.RoomID) error {
	return b.typing.StopTyping(ctx, roomID)
}

// ----- BotSync

func (b *DefaultBot) StartListening(ctx context.Context) error {
	return b.sync.StartListening(ctx)
}

func (b *DefaultBot) StopListening(ctx context.Context) error {
	return b.sync.StopListening(ctx)
}

// ----- BotHealth

func (b *DefaultBot) Ping(ctx context.Context) error {
	return b.health.Ping(ctx)
}

// ----- BotMedia

func (b *DefaultBot) Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error) {
	return b.media.Download(ctx, mxcURL)
}

// ----- BotCrypto

func (b *DefaultBot) DecryptEvent(ctx context.Context, evt *event.Event) (*event.Event, error) {
	return b.crypto.DecryptEvent(ctx, evt)
}

func (b *DefaultBot) Encrypt(
	ctx context.Context,
	roomID id.RoomID,
	eventType event.Type,
	content any,
) (event.Type, any, error) {
	return b.crypto.Encrypt(ctx, roomID, eventType, content)
}

func (b *DefaultBot) IsEncryptedRoom(ctx context.Context, roomID id.RoomID) (bool, error) {
	return b.crypto.IsEncryptedRoom(ctx, roomID)
}

func (b *DefaultBot) EnsureOutboundSession(ctx context.Context, roomID id.RoomID) error {
	return b.crypto.EnsureOutboundSession(ctx, roomID)
}

func (b *DefaultBot) IsEncrypted(evt *event.Event) bool {
	return b.crypto.IsEncrypted(evt)
}

func (b *DefaultBot) RequestKey(ctx context.Context, evt *event.Event) error {
	return b.crypto.RequestKey(ctx, evt)
}

func (b *DefaultBot) HandleToDevice(ctx context.Context, evt *event.Event) {
	b.crypto.HandleToDevice(ctx, evt)
}

// ----- EventDispatcher

func (b *DefaultBot) AddEventHandler(h handlers.EventHandler) {
	b.dispatcher.AddEventHandler(h)
}

func (b *DefaultBot) AddFilter(f filters.Filter) {
	b.dispatcher.AddFilter(f)
}
