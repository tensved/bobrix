package ctx

import (
	"context"
	"log/slog"

	dbot "github.com/tensved/bobrix/mxbot/domain/bot"
	dctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	"maunium.net/go/mautrix/event"
)

func injectMetadataInContext(
	ctx context.Context,
	evt *event.Event,
	loader dbot.EventLoader,
) context.Context {

	meta := map[string]any{
		"event": evt,
	}

	if evt == nil {
		slog.Warn("event is nil, skipping metadata injection")
		return context.WithValue(ctx, dctx.MetadataKeyContext, meta)
	}

	msg := evt.Content.AsMessage()
	if msg == nil || msg.RelatesTo == nil {
		return context.WithValue(ctx, dctx.MetadataKeyContext, meta)
	}

	meta["thread_id"] = msg.RelatesTo.EventID

	if loader != nil {
		main, err := loader.GetEvent(ctx, evt.RoomID, msg.RelatesTo.EventID)
		if err != nil {
			slog.Error("get main event failed", "err", err)
		}
		if main != nil {
			if v, ok := main.Content.Raw[dctx.AnswerToCustomField]; ok {
				meta["thread.answer_to"] = v
			}
		} else {
			slog.Warn("main event is nil, skipping thread.answer_to")
		}
	}

	return context.WithValue(ctx, dctx.MetadataKeyContext, meta)
}
