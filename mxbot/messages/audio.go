package messages

import (
	"fmt"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"time"
)

// NewAudio - creates a new audio message
// Audio - audio bytes
// Text - message text (optional: take first argument if set). Default: audio_YYYY-MM-DD_HH-MM-SS.mp3
func NewAudio(audio []byte, text ...string) Message {

	t := fmt.Sprintf("audio_%s.mp3", time.Now().Format(time.RFC3339))

	if len(text) > 0 {
		t = text[0]
	}

	return &BaseMessage{
		msgType: event.MsgAudio,
		text:    "Voice message",
		file: &FileInfo{
			fileName:     t,
			mimeType:     "audio/mpeg",
			contentBytes: audio,

			contentURI: id.ContentURI{},
		},
	}
}
