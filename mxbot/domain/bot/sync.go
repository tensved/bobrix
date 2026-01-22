package bot

import "context"

type BotSync interface {
	StartListening(ctx context.Context) error
	StopListening(ctx context.Context) error
}
