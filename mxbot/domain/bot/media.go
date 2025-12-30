package bot

import (
	"context"

	"maunium.net/go/mautrix/id"
)

type BotMedia interface {
	Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error)
}
