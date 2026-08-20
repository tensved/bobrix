package bobrix

import (
	"encoding/json"
	"testing"

	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// parsedEvent builds an event whose Content.Parsed is already populated, the same
// precondition InsertRoomHistoryToContextHandler guarantees (it calls ParseRaw itself)
// before ever handing a thread to ConvertThreadToMessages — which, unlike that handler,
// does not parse content itself and expects it done already.
func parsedEvent(t *testing.T, sender string, evtType event.Type, c any) *event.Event {
	t.Helper()
	evt := &event.Event{
		Sender:  id.UserID(sender),
		Type:    evtType,
		Content: event.Content{VeryRaw: rawContent(t, c)},
	}
	if err := evt.Content.ParseRaw(evtType); err != nil {
		t.Fatalf("failed to parse content: %v", err)
	}
	return evt
}

func rawContent(t *testing.T, c any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal content: %v", err)
	}
	return b
}

func textEvt(t *testing.T, sender, body string) *event.Event {
	t.Helper()
	return parsedEvent(t, sender, event.EventMessage, event.MessageEventContent{MsgType: event.MsgText, Body: body})
}

func imageEvt(t *testing.T, sender, url string) *event.Event {
	t.Helper()
	return parsedEvent(t, sender, event.EventMessage, event.MessageEventContent{
		MsgType: event.MsgImage,
		URL:     id.ContentURIString(url),
	})
}

func TestConvertThreadToMessages_AssignsRoleByBotSender(t *testing.T) {
	thread := &mxbot.MessagesThread{
		Messages: []*event.Event{
			textEvt(t, "@user:hs", "hi bot"),
			textEvt(t, "@bot:hs", "hi there"),
		},
	}

	got := ConvertThreadToMessages(thread, "@bot:hs")

	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}
	if got[0].Role != contracts.UserRole || got[0].Content != "hi bot" {
		t.Fatalf("expected first message from the user, got %+v", got[0])
	}
	if got[1].Role != contracts.AssistantRole || got[1].Content != "hi there" {
		t.Fatalf("expected second message from the bot, got %+v", got[1])
	}
}

func TestConvertThreadToMessages_CarriesAttachmentURL(t *testing.T) {
	thread := &mxbot.MessagesThread{
		Messages: []*event.Event{
			imageEvt(t, "@user:hs", "mxc://hs/abc123"),
		},
	}

	got := ConvertThreadToMessages(thread, "@bot:hs")

	if len(got) != 1 {
		t.Fatalf("expected 1 message, got %d", len(got))
	}
	if len(got[0].Attachments) != 1 || got[0].Attachments[0] != "mxc://hs/abc123" {
		t.Fatalf("expected the mxc:// URL carried through as an attachment, got %+v", got[0].Attachments)
	}
	if got[0].Content != "" {
		t.Fatalf("expected no text body on a bare image message, got %q", got[0].Content)
	}
}

func TestConvertThreadToMessages_TextMessageHasNoAttachments(t *testing.T) {
	thread := &mxbot.MessagesThread{
		Messages: []*event.Event{textEvt(t, "@user:hs", "just text")},
	}

	got := ConvertThreadToMessages(thread, "@bot:hs")

	if len(got) != 1 || len(got[0].Attachments) != 0 {
		t.Fatalf("expected a plain text message with no attachments, got %+v", got)
	}
}

func TestConvertThreadToMessages_UnparsableContentIsSkipped(t *testing.T) {
	thread := &mxbot.MessagesThread{
		Messages: []*event.Event{
			// StateTombstone is a registered type but not a message shape: its Content
			// parses to a *TombstoneEventContent, so AsMessage()'s type assertion to
			// *MessageEventContent fails and falls back to a zero value with neither
			// a body nor an attachment.
			parsedEvent(t, "@user:hs", event.StateTombstone, event.TombstoneEventContent{Body: "upgraded"}),
			textEvt(t, "@user:hs", "still here"),
		},
	}

	got := ConvertThreadToMessages(thread, "@bot:hs")

	if len(got) != 1 || got[0].Content != "still here" {
		t.Fatalf("expected the unparsable event dropped and the text message kept, got %+v", got)
	}
}

func TestConvertThreadToMessages_EmptyThread(t *testing.T) {
	got := ConvertThreadToMessages(&mxbot.MessagesThread{}, "@bot:hs")

	if len(got) != 0 {
		t.Fatalf("expected no messages for an empty thread, got %+v", got)
	}
}
