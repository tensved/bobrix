package messages

import (
	"fmt"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"mime"
)

// NewFile - creates a new file message
// Bytes - file bytes
// Name - file fileName
// FileType - file type
// Text - message text (optional: take first argument if set). Default: file_YYYY-MM-DD_HH-MM-SS.png
func NewFile(bytes []byte, name string, fileType string, text ...string) Message {
	t := fmt.Sprintf("file_%s.%s", name, fileType)

	if len(text) > 0 {
		t = text[0]
	}

	mimeType := mime.TypeByExtension(fileType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &BaseMessage{
		text:    t,
		msgType: event.MsgFile,
		file: &FileInfo{
			fileName:     name,
			mimeType:     mimeType,
			contentBytes: bytes,
			contentURI:   id.ContentURI{},
		},
	}
}
