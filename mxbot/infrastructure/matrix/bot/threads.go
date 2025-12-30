package bot

import (
	"context"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"maunium.net/go/mautrix/event"
)

var _ domainbot.Threads = (*DefaultBot)(nil)

func (b *DefaultBot) IsThreadEnabled() bool {
	return b.isThreadEnabled
}

func (b *DefaultBot) GetThreadByEvent(ctx context.Context, evt *event.Event) (*threads.MessagesThread, error) {
	if evt == nil {
		return nil, ErrNilEvent
	}

	rel := evt.Content.AsMessage().RelatesTo

	if rel == nil || rel.Type != event.RelThread {
		return nil, nil
	}

	roomID := evt.RoomID
	parentEventID := rel.EventID

	return b.GetThread(ctx, roomID, parentEventID)
}
