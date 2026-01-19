package mxbot

import "maunium.net/go/mautrix"

func TextCommand(
	command string,
	handler func(Ctx) error,
) EventHandler {
	return NewMessageHandler(
		func(ctx Ctx) error {
			body := ctx.Event().Content.AsMessage().Body
			if body != command {
				return nil
			}
			return handler(ctx)
		},
		FilterMessageText(),
	)
}

func AsMatrixClient(raw any) (*mautrix.Client, bool) {
	c, ok := raw.(*mautrix.Client)
	return c, ok
}
