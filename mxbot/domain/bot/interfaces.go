package bot

// BotMessaging - can send messages and join rooms
// type BotMessaging interface {
// 	SendMessage(ctx context.Context, roomID id.RoomID, msg messages.Message) error
// 	JoinRoom(ctx context.Context, roomID id.RoomID) error
// }

// // BotInfo - bot identity
// type BotInfo interface {
// 	UserID() id.UserID
// 	FullName() string
// 	Name() string
// }

// type BotThreads interface {
// 	IsThreadEnabled() bool
// 	GetThreadByEvent(ctx context.Context, evt *event.Event) (*threads.MessagesThread, error)
// }

// // Role-based composition
// type BotJoiner interface {
// 	BotMessaging
// 	BotInfo
// }
