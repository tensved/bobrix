package bot

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	threads "github.com/tensved/bobrix/mxbot/domain/threads"
)

// inner helper infrastructure method
// GetThreadStory - gets the thread story
func (b *DefaultBot) GetThread(ctx context.Context, roomID id.RoomID, parentEventID id.EventID) (*threads.MessagesThread, error) {

	msgs, err := b.matrixClient.Messages(
		ctx,
		roomID,
		"",
		"",
		mautrix.DirectionBackward,
		nil,
		b.credentials.ThreadLimit,
	)
	if err != nil {
		slog.Error("error get messages", "error", err)
		return nil, err
	}

	data := make([]*event.Event, 0)

	for _, evt := range msgs.Chunk {

		msg, ok := GetFixedMessage(evt)
		if !ok {
			continue
		}

		evt.Content.Parsed = msg

		if evt.ID == parentEventID {

			data = append(data, evt)

			rawContent := evt.Content.Raw

			if answerEventID, ok := rawContent[AnswerToCustomField]; ok {

				evtID := id.EventID(answerEventID.(string))

				answerEvent, err := b.matrixClient.GetEvent(ctx, roomID, evtID)
				if err != nil {
					slog.Error("error get answer event", "error", err)
					break
				}

				answerMsg, ok := GetFixedMessage(answerEvent)
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

func GetFixedMessage(evt *event.Event) (*event.MessageEventContent, bool) {
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
