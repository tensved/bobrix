package mxbot

import (
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func TextCommand(
	command string,
	handler func(Ctx) error,
) EventHandler {
	return NewMessageHandler(
		handler,
		func(evt *event.Event) bool {
			msg := evt.Content.AsMessage()
			if msg == nil {
				return false
			}

			body := strings.TrimSpace(msg.Body)

			return body == command
		},
		FilterMessageText(),
	)
}

func AsMatrixClient(raw any) (*mautrix.Client, bool) {
	c, ok := raw.(*mautrix.Client)
	return c, ok
}
