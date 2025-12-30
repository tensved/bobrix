package middleware

import (
	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/contracts"
)

func DecryptMiddleware(crypto domainbot.Crypto) Middleware {
	return func(next EventHandler) EventHandler {
		return contracts.HandlerFunc(func(ctx ctx.Ctx) error {
			evt := ctx.Event()

			if crypto.IsEncrypted(evt) {
				decrypted, err := crypto.Decrypt(ctx.Context(), evt)
				if err != nil {
					return err
				}
				ctx.SetEvent(decrypted)
			}

			return next.Handle(ctx)
		})
	}
}
