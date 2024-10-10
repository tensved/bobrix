package messages

import (
	"encoding/json"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type BaseMessage struct {
	msgType      event.MessageType
	text         string
	contentURI   id.ContentURI
	customFields map[string]any
}

func (bm *BaseMessage) Type() event.MessageType {
	return bm.msgType
}

func (bm *BaseMessage) AsEvent(_ *event.RelatesTo) event.MessageEventContent {
	slog.Info("AsEvent", "type", bm.Type())
	return event.MessageEventContent{
		MsgType: bm.Type(),
		Body:    bm.text,
	}
}

func (bm *BaseMessage) AsJSON() map[string]any {

	evt := bm.AsEvent(nil)
	b, err := json.Marshal(evt)
	if err != nil {
		panic(err)
	}

	var structMap map[string]any

	if err := json.Unmarshal(b, &structMap); err != nil {
		panic(err)
	}

	for key, value := range bm.customFields {
		structMap[key] = value
	}

	return structMap
}

func (bm *BaseMessage) AsReqUpload() mautrix.ReqUploadMedia {
	return mautrix.ReqUploadMedia{}
}

func (bm *BaseMessage) SetContentURI(contentURI id.ContentURI) {
	bm.contentURI = contentURI
}

func (bm *BaseMessage) AddCustomFields(values ...any) {

	if len(values) == 0 {
		return
	}

	if len(values)%2 != 0 {
		panic("invalid key-value pair: missing key or value")
	}

	for i := 0; i < len(values)-1; i += 2 {

		key := values[i].(string)
		val := values[i+1]

		bm.customFields[key] = val
	}

}

func NewBaseMessage(text string) BaseMessage {
	return BaseMessage{
		text:         text,
		customFields: make(map[string]any),
	}
}

// Message - message interface
// should be implemented by all message types
// It is used to send messages to the room by the bot
type Message interface {
	Type() event.MessageType                                // get message type
	AsEvent(rel *event.RelatesTo) event.MessageEventContent // get event contentBytes
	AsJSON() map[string]any                                 // get message as JSON
	AsReqUpload() mautrix.ReqUploadMedia                    // get upload media request
	SetContentURI(contentURI id.ContentURI)                 // set contentBytes URI after upload

	AddCustomFields(values ...any) // add custom fields to message. Required use as < key, value, ... >
}

type Options func(m Message)

/*
Есть стурктура BaseMessage
Есть много функций, которые создают Message (NewText, NewAudio)
т.е. самих структур TextMessage - нет.
*/
