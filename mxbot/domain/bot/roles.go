package bot

type FullBot interface {
	BotAuth
	BotInfo
	BotMessaging
	BotThreads
	BotCrypto
	BotClient
	EventLoader
	BotRoomActions
	BotTyping
	BotSync
	EventDispatcher
	BotHealth
	BotPresenceControl
	BotMedia
}
