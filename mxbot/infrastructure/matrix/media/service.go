package media

import (
	"context"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
)

var _ dbot.BotMedia = (*Service)(nil)

type Service struct {
	client *mautrix.Client
}

func New(c dbot.BotClient) *Service {
	return &Service{
		client: c.RawClient().(*mautrix.Client),
	}
}

// Download - downloads the content of the mxc URL
func (s *Service) Download(ctx context.Context, mxcURL id.ContentURI) ([]byte, error) {
	return s.client.DownloadBytes(ctx, mxcURL)
}
