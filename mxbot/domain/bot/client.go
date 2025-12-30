package bot

// only for infra
type BotClient interface {
	RawClient() any
}
