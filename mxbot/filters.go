package mxbot

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"strings"
	"time"
)

// Filter - message filter
// return true if message should be processed
// return false if message should be ignored
type Filter func(evt *event.Event) bool

// FilterNotMe - filter for messages from other users
// (ignores messages from the bot itself)
func FilterNotMe(matrixClient *mautrix.Client) Filter {
	return func(evt *event.Event) bool {
		return evt.Sender != matrixClient.UserID
	}
}

// FilterAfterStart - filter for messages after start time
// (ignores messages that were sent before start time)
func FilterAfterStart(startTime time.Time) Filter {
	return func(evt *event.Event) bool {
		eventTime := time.UnixMilli(evt.Timestamp)

		return eventTime.After(startTime)
	}
}

// FilterCommand - filter for command messages
// (check if message starts with command prefix and name)
func FilterCommand(command *Command) Filter {
	return func(evt *event.Event) bool {

		if evt.Type != event.EventMessage {
			return false
		}

		msg := evt.Content.AsMessage().Body
		wordsInMessage := strings.Split(msg, " ")

		if len(wordsInMessage) < 1 {
			return false
		}

		commandPrefix := command.Prefix + command.Name
		return strings.EqualFold(wordsInMessage[0], commandPrefix)
	}
}
