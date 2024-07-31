package ada

import (
	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
	"log/slog"
)

func NewAdaBot(username string, password string, homeServerURL string) (*bobrix.Bobrix, error) {
	bot, err := mxbot.NewDefaultBot("ada",
		&mxbot.BotCredentials{
			Username:      username,
			Password:      password,
			HomeServerURL: homeServerURL,
		})
	if err != nil {
		return nil, err
	}
	//
	bot.AddCommand(mxbot.NewCommand(
		"ping",
		func(c mxbot.CommandCtx) error {

			return c.Answer("pong")
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

		err := ctx.Answer(answer)

		if err != nil {
			slog.Error("failed to send message", "error", err)
		}

	})

	return bobr, nil
}
