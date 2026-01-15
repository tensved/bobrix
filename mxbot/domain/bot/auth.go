package bot

import "context"

type BotAuth interface {
	Authorize(ctx context.Context) error
}
