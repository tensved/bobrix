package messages

import "maunium.net/go/mautrix/event"

func NewText(text string) Message {
	return &BaseMessage{
		msgType:         event.MsgText,
		text:            text,
		markDownSupport: MarkDownSupportDefault,
	}
}
