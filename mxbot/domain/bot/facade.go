package bot

type FullBot interface {
	BotAuth
	BotInfo
	// BotMessaging // sens messages only using ctx.Answer (application lvl)
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
