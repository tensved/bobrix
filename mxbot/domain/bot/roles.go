package bot // ok

// type Joiner interface {
// 	Messaging
// 	Info
// }

type FullBot interface {
	Info
	BotMessaging
	Threads
	Crypto
}
