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

	bot.AddEventHandler(
		mxbot.NewLoggerHandler("ada"),
	)

	bobr := bobrix.NewBobrix(bot, bobrix.WithHealthcheck(bobrix.WithAutoSwitch()))

	bobr.SetContractParser(bobrix.DefaultContractParser())

	bobr.SetContractParser(bobrix.AutoRequestParser(&bobrix.AutoParserOpts{
		ServiceName: "ada",
		MethodName:  "generate",
		InputName:   "prompt",
	}))

	bobr.ConnectService(NewADAService("hilltwinssl.singularitynet.io"), func(ctx mxbot.Ctx, r *contracts.MethodResponse) {

		if r.Err != nil {
			slog.Error("failed to process request", "error", r.Err)

			if err := ctx.TextAnswer("error: " + r.Err.Error()); err != nil {
				slog.Error("failed to send message", "error", err)
			}

			return
		}

		answer, ok := r.GetString("answer")
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
