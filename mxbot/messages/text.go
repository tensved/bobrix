package messages

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var _ Message = (*Text)(nil)

type Text struct {
	text string
}

func (m *Text) Type() event.MessageType {
	return event.MsgText
}

func (m *Text) AsEvent() event.MessageEventContent {
	return event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    m.text,
	}
}

func (m *Text) AsReqUpload(_ id.RoomID) mautrix.ReqUploadMedia {
	return mautrix.ReqUploadMedia{}
}

func (m *Text) SetContentURI(_ id.ContentURI) {}

func NewText(text string) *Text {
	return &Text{
		text: text,
	}
}
