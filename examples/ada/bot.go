package ada

import (
	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
	"log/slog"
)

func NewAdaBot(credentials *mxbot.BotCredentials) (*bobrix.Bobrix, error) {
	bot, err := mxbot.NewDefaultBot("ada", credentials)
	if err != nil {
		return nil, err
	}
	//
	bot.AddCommand(mxbot.NewCommand(
		"ping",
		func(c mxbot.CommandCtx) error {

			return c.TextAnswer("pong")
		}),
	)

	bot.AddEventHandler(
		mxbot.AutoJoinRoomHandler(bot),
	)

	bobr := bobrix.NewBobrix(bot)

	bobr.SetContractParser(bobrix.DefaultContractParser())

	bobr.ConnectService(NewADAService("hilltwinssl.singularitynet.io"), func(ctx mxbot.Ctx, r *contracts.MethodResponse) {

		answer, ok := r.Data["answer"].(string)
		if !ok {
			answer = "I don't know"
		}

		err := ctx.TextAnswer(answer)

		if err != nil {
			slog.Error("failed to send message", "error", err)
		}

	})

	return bobr, nil
}
