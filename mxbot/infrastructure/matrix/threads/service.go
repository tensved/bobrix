package threads

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

	domain "github.com/tensved/bobrix/mxbot/domain/bot"
	dctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var _ domain.BotThreads = (*Service)(nil)

type Service struct {
	client          *mautrix.Client
	isThreadEnabled bool
	threadLimit     int
}

func New(c domain.BotClient, isThreadEnabled bool, threadLimit int) *Service {
	return &Service{
		client:          c.RawClient().(*mautrix.Client),
		isThreadEnabled: isThreadEnabled,
		threadLimit:     threadLimit,
	}
}

func (s *Service) IsThreadEnabled() bool {
	return s.isThreadEnabled
}

func (s *Service) GetThreadByEvent(ctx context.Context, evt *event.Event) (*threads.MessagesThread, error) {
	if evt == nil {
		return nil, domain.ErrNilEvent
	}

	rel := evt.Content.AsMessage().RelatesTo

	if rel == nil || rel.Type != event.RelThread {
		return nil, nil
	}

	roomID := evt.RoomID
	parentEventID := rel.EventID

	return s.GetThread(ctx, roomID, parentEventID)
}

// inner helper infrastructure method
// GetThreadStory - gets the thread story
func (s *Service) GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*threads.MessagesThread, error) {
	msgs, err := s.client.Messages(
		ctx,
		roomID,
		"",
		"",
		mautrix.DirectionBackward,
		nil,
		s.threadLimit,
	)
	if err != nil {
		slog.Error("error get messages", "error", err)
		return nil, err
	}

	data := make([]*event.Event, 0)

	for _, evt := range msgs.Chunk {

		msg, ok := getFixedMessage(evt)
		if !ok {
			continue
		}

		evt.Content.Parsed = msg

		if evt.ID == parentEventID {

			data = append(data, evt)

			rawContent := evt.Content.Raw

			if answerEventID, ok := rawContent[dctx.AnswerToCustomField]; ok {

				evtID := id.EventID(answerEventID.(string))

				answerEvent, err := s.client.GetEvent(ctx, roomID, evtID)
				if err != nil {
					slog.Error("error get answer event", "error", err)
					break
				}

				answerMsg, ok := getFixedMessage(answerEvent)
				if !ok {
					continue
				}

				answerEvent.Content.Parsed = answerMsg

				data = append(data, answerEvent)
			}
			break
		}

		rel := msg.RelatesTo

		if rel == nil {
			slog.Info("rel type nil", "event_id", evt.ID)
			continue
		}

		if rel.Type != event.RelThread || rel.EventID != parentEventID {
			slog.Debug("rel type not thread", "event_id", evt.ID, "rel_event_id", rel.EventID, "rel_type", rel.Type)
			continue
		}

		data = append(data, evt)
	}

	slog.Info("data size", "size", len(data))

	slices.Reverse(data)

	thread := &threads.MessagesThread{
		Messages: data,
		ParentID: parentEventID,
		RoomID:   roomID,
	}

	return thread, nil
}

func getFixedMessage(evt *event.Event) (*event.MessageEventContent, bool) {
	veryRaw := evt.Content.VeryRaw

	if veryRaw == nil {
		return nil, false
	}

	var msg *event.MessageEventContent

	if err := json.Unmarshal(veryRaw, &msg); err != nil {
		slog.Error("error unmarshal message", "error", err)
		return nil, false
	}

	return msg, true
}
