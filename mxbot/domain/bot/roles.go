package bot

type FullBot interface {
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
}
