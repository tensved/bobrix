package messages

import (
	"fmt"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"mime"
)

var _ Message = (*File)(nil)

type File struct {
	contentBytes []byte
	name         string
	fileType     string
	contentURI   id.ContentURI
	text         string
}

func (m *File) Type() event.MessageType {
	return event.MsgFile
}

func (m *File) AsEvent() event.MessageEventContent {

	content := event.MessageEventContent{
		Body:    m.text,
		MsgType: event.MsgFile,
		URL:     m.contentURI.CUString(),
		Info: &event.FileInfo{
			MimeType: m.fileType,
		},
	}

	return content
}

func (m *File) AsReqUpload() mautrix.ReqUploadMedia {

	return mautrix.ReqUploadMedia{
		ContentBytes: m.contentBytes,
		ContentType:  m.fileType,
		FileName:     m.name,
	}
}

func (m *File) SetContentURI(contentURI id.ContentURI) {
	m.contentURI = contentURI
}

// NewFile - creates a new file message
// Bytes - file bytes
// Name - file name
// FileType - file type
// Text - message text (optional: take first argument if set). Default: file_YYYY-MM-DD_HH-MM-SS.png
func NewFile(bytes []byte, name string, fileType string, text ...string) *File {
	t := fmt.Sprintf("file_%s.%s", name, fileType)

	if len(text) > 0 {
		t = text[0]
	}

	mimeType := mime.TypeByExtension(fileType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &File{
		contentBytes: bytes,
		text:         t,
		name:         name,
		fileType:     mimeType,
	}
}