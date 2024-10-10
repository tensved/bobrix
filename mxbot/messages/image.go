package messages

import (
	"fmt"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"time"
)

var _ Message = (*Image)(nil)

type Image struct {
	image []byte
	BaseMessage
}

func (m *Image) Type() event.MessageType {
	return event.MsgImage
}

func (m *Image) AsEvent(_ *event.RelatesTo) event.MessageEventContent {
	content := event.MessageEventContent{
		Body:    m.text,
		MsgType: m.Type(),
		URL:     m.contentURI.CUString(),
		Info: &event.FileInfo{
			MimeType: "image/png",
		},
	}

	return content
}

func (m *Image) AsReqUpload() mautrix.ReqUploadMedia {
	return mautrix.ReqUploadMedia{
		ContentBytes: m.image,
		ContentType:  "image/png",
		FileName:     m.text,
	}
}

// NewImage - creates a new image message
// Image - image bytes (base64 encoded)
// Text - message text (optional: take first argument if set). Default: image_YYYY-MM-DD_HH-MM-SS.png
func NewImage(image []byte, text ...string) *Image {
	t := fmt.Sprintf("image_%s.png", time.Now().Format(time.RFC3339))

	if len(text) > 0 {
		t = text[0]
	}

	bm := NewBaseMessage(t)

	return &Image{
		image:       image,
		BaseMessage: bm,
	}
}
