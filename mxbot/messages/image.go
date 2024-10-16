package messages

import (
	"fmt"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"time"
)

// NewImage - creates a new image message
// Image - image bytes (base64 encoded)
// Text - message text (optional: take first argument if set). Default: image_YYYY-MM-DD_HH-MM-SS.png
func NewImage(image []byte, text ...string) Message {
	t := fmt.Sprintf("image_%s.png", time.Now().Format(time.RFC3339))

	if len(text) > 0 {
		t = text[0]
	}

	return &BaseMessage{
		text:    t,
		msgType: event.MsgImage,
		file: &FileInfo{
			contentBytes: image,
			mimeType:     "image/png",
			contentURI:   id.ContentURI{},
		},
	}
}
