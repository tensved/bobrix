package messages

import (
	"encoding/json"
	"github.com/gomarkdown/markdown"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var MarkDownSupportDefault bool

func init() {
	MarkDownSupportDefault = true
}

type Message interface {
	Type() event.MessageType                // get message type
	AsEvent() event.MessageEventContent     // get event contentBytes
	AsJSON() map[string]any                 // get message as JSON
	AsReqUpload() mautrix.ReqUploadMedia    // get upload media request
	SetContentURI(contentURI id.ContentURI) // set contentBytes URI after upload

	SetRelatesTo(rel *event.RelatesTo)

	AddCustomFields(values ...any) // add custom fields to message. Required use as < key, value, ... >

	SetMarkDownSupport(status bool) // set markDown support
}

type FileInfo struct {
	fileName string
	mimeType string

	contentBytes []byte
	contentURI   id.ContentURI
}

var _ Message = (*BaseMessage)(nil)

type BaseMessage struct {
	msgType event.MessageType

	rel *event.RelatesTo

	text string

	markDownSupport bool

	file *FileInfo

	customFields map[string]any
}

func (m *BaseMessage) Type() event.MessageType {
	return m.msgType
}

func (m *BaseMessage) AsEvent() event.MessageEventContent {
	evt := event.MessageEventContent{
		MsgType: m.Type(),
		Body:    m.text,
	}

	if m.markDownSupport {

		formattedBody := string(markdown.ToHTML([]byte(m.text), nil, nil))

		evt.Format = "org.matrix.custom.html"
		evt.FormattedBody = formattedBody
	}

	if m.rel != nil {
		evt.RelatesTo = m.rel
	}

	if m.file != nil {
		evt.URL = m.file.contentURI.CUString()
		evt.Info = &event.FileInfo{
			MimeType: m.file.mimeType,
		}
	}

	return evt
}

func (m *BaseMessage) AsJSON() map[string]any {
	var result = make(map[string]any)

	d, err := json.Marshal(m.AsEvent())
	if err != nil {
		slog.Error("failed to marshal message", "error", err)
		return nil
	}

	err = json.Unmarshal(d, &result)
	if err != nil {
		slog.Error("failed to unmarshal message", "error", err)
		return nil
	}

	if m.customFields != nil {
		for k, v := range m.customFields {
			result[k] = v
		}
	}

	return result
}

func (m *BaseMessage) AsReqUpload() mautrix.ReqUploadMedia {
	return mautrix.ReqUploadMedia{
		ContentBytes: m.file.contentBytes,
		ContentType:  m.file.mimeType,
		FileName:     m.file.fileName,
	}
}

func (m *BaseMessage) SetRelatesTo(rel *event.RelatesTo) {
	m.rel = rel
}

func (m *BaseMessage) SetContentURI(contentURI id.ContentURI) {
	m.file.contentURI = contentURI
}

func (m *BaseMessage) AddCustomFields(values ...any) {
	if m.customFields == nil {
		m.customFields = map[string]any{}
	}
	for i := 0; i < len(values); i += 2 {
		m.customFields[values[i].(string)] = values[i+1]
	}
}

func (m *BaseMessage) SetMarkDownSupport(status bool) {
	m.markDownSupport = status
}
