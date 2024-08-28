package messages

import (
	"fmt"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"time"
)

var _ Message = (*Audio)(nil)

type Audio struct {
	audio      []byte
	contentURI id.ContentURI
	text       string
}

func (m *Audio) Type() event.MessageType {
	return event.MsgAudio
}

func (m *Audio) AsEvent() event.MessageEventContent {
	content := event.MessageEventContent{
		Body:    "Voice message",
		MsgType: event.MsgAudio,
		URL:     m.contentURI.CUString(),
		Info: &event.FileInfo{
			MimeType: "audio/mpeg",
			Size:     len(m.audio),
		},
	}

	return content
}

func (m *Audio) AsReqUpload() mautrix.ReqUploadMedia {

	return mautrix.ReqUploadMedia{
		ContentBytes:  m.audio,
		ContentType:   "audio/mpeg",
		ContentLength: int64(len(m.audio)),
		FileName:      m.text,
	}
}

func (m *Audio) SetContentURI(contentURI id.ContentURI) {
	m.contentURI = contentURI
}

// NewAudio - creates a new audio message
// Audio - audio bytes
// Text - message text (optional: take first argument if set). Default: audio_YYYY-MM-DD_HH-MM-SS.mp3
func NewAudio(audio []byte, text ...string) *Audio {
	t := fmt.Sprintf("audio_%s.mp3", time.Now().Format(time.RFC3339))

	if len(text) > 0 {
		t = text[0]
	}

	return &Audio{
		audio: audio,
		text:  t,
	}
}
