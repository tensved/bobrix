package filters

import (
	"slices"

	df "github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix/event"
)

// FilterMessageTypes - filter for specific message types
// check if message type is in the list of message types (event.MessageType)
// also check if event type is event.EventMessage
// return true if message type is in the list
func FilterMessageTypes(types ...event.MessageType) df.Filter {
	return func(evt *event.Event) bool {
		if evt.Type != event.EventMessage {
			return false
		}
		return slices.Contains(types, evt.Content.AsMessage().MsgType)
	}
}

// FilterMessageText - filter for text messages
// check if message type is event.MsgText
// return true if message type is event.MsgText
func FilterMessageText() df.Filter {
	return FilterMessageTypes(event.MsgText)
}

// FilterMessageAudio - filter for audio messages
// check if message type is event.MsgAudio
// return true if message type is event.MsgAudio
func FilterMessageAudio() df.Filter {
	return FilterMessageTypes(event.MsgAudio)
}

// FilterMessageImage - filter for image messages
// check if message type is event.MsgImage
// return true if message type is event.MsgImage
func FilterMessageImage() df.Filter {
	return FilterMessageTypes(event.MsgImage)
}

// FilterMessageVideo - filter for video messages
// check if message type is event.MsgVideo
// return true if message type is event.MsgVideo
func FilterMessageVideo() df.Filter {
	return FilterMessageTypes(event.MsgVideo)
}

// FilterMessageFile - filter for file messages
// check if message type is event.MsgFile
// return true if message type is event.MsgFile
func FilterMessageFile() df.Filter {
	return FilterMessageTypes(event.MsgFile)
}
