package messages

import (
	"encoding/json"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var _ Message = (*Text)(nil)

type Text struct {
	text         string
	customFields map[string]any
}

func (m *Text) Type() event.MessageType {
	return event.MsgText
}

func (m *Text) AsEvent(rel *event.RelatesTo) event.MessageEventContent {
	return event.MessageEventContent{
		Body:      m.text,
		MsgType:   m.Type(),
		RelatesTo: rel,
	}
}

func (m *Text) AsJSON() map[string]any {
	var data = make(map[string]any)

	d, err := json.Marshal(m.AsEvent(nil))

	if err != nil {
		return nil
	}

	err = json.Unmarshal(d, &data)

	if err != nil {
		return nil
	}

	slog.Info("as json", "data", data)

	return data
}

func (m *Text) AsReqUpload() mautrix.ReqUploadMedia {
	return mautrix.ReqUploadMedia{}
}

func (m *Text) SetContentURI(_ id.ContentURI) {
	return
}

func (m *Text) AddCustomFields(values ...any) {
	for i := 0; i < len(values); i += 2 {
		m.customFields[values[i].(string)] = values[i+1]
	}
}

func NewText(text string) *Text {
	return &Text{
		text:         text,
		customFields: make(map[string]any),
	}
}
