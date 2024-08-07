package messages

import (
	"fmt"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"time"
)

var _ Message = (*Image)(nil)

type Image struct {
	image      []byte
	contentURI id.ContentURI
	text       *string
}

func (m *Image) Type() event.MessageType {
	return event.MsgImage
}

func (m *Image) AsEvent() event.MessageEventContent {
	content := event.MessageEventContent{
		MsgType: event.MsgImage,
		URL:     m.contentURI.CUString(),
		Info: &event.FileInfo{
			MimeType: "image/png",
		},
	}

	if m.text != nil {
		content.Body = *m.text
	}

	return content
}

func (m *Image) AsReqUpload(roomID id.RoomID) mautrix.ReqUploadMedia {
	return mautrix.ReqUploadMedia{
		ContentBytes: m.image,
		ContentType:  "image/png",
		FileName:     fmt.Sprintf("%s_%s.png", roomID, time.Now().Format(time.RFC3339)),
	}
}

func (m *Image) SetContentURI(contentURI id.ContentURI) {
	m.contentURI = contentURI
}

func NewImage(image []byte, text ...string) *Image {
	var t *string

	if len(text) > 0 {
		t = &text[0]
	}

	return &Image{
		image: image,
		text:  t,
	}
}
