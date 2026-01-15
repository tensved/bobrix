package ada

import (
	"log/slog"

	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
)

func NewAdaBot(cfg *mxbot.Config) (*bobrix.Bobrix, error) {
	// создаём Matrix-бота (инфраструктура)
	bot, err := mxbot.NewMatrixBot(*cfg)
	if err != nil {
		return nil, err
	}

	// Bobrix — application слой
	bx := bobrix.NewBobrix(bot, bobrix.WithHealthcheck(bobrix.WithAutoSwitch()))

	// --- контрактный парсер
	bx.SetContractParser(
		bobrix.DefaultContractParser(bot),
	)

	// --- авто-парсер (любое сообщение → ada.generate)
	bx.SetContractParser(
		bobrix.AutoRequestParser(&bobrix.AutoParserOpts{
			Bot:         bot,
			ServiceName: "ada",
			MethodName:  "generate",
			InputName:   "prompt",
		}),
	)

	// --- подключаем сервис
	bx.ConnectService(NewADAService("hilltwinssl.singularitynet.io"),
		func(ctx mxbot.Ctx, r *contracts.MethodResponse, _ any) {

			if r.Err != nil {
				slog.Error("ada failed", "error", r.Err)
				_ = ctx.TextAnswer("error: " + r.Err.Error())
				return
			}

			answer, ok := r.GetString("text")
			if !ok {
				answer = "I don't know"
			}

			slog.Debug("ada response", "text", answer)
			_ = ctx.TextAnswer(answer)
		},
	)

	return bx, nil
}

// package ada

// import (
// 	"github.com/tensved/bobrix"
// 	"github.com/tensved/bobrix/contracts"
// 	"github.com/tensved/bobrix/mxbot"
// 	"log/slog"
// )

// func NewAdaBot(credentials *mxbot.Config) (*bobrix.Bobrix, error) {
// 	bot, err := mxbot.NewDefaultBot("ada", credentials)
// 	if err != nil {
// 		return nil, err
// 	}

// 	bot.AddCommand(mxbot.NewCommand(
// 		"ping",
// 		func(c mxbot.CommandCtx) error {

// 			return c.TextAnswer("pong")
// 		}),
// 	)

// 	bot.AddEventHandler(
// 		mxbot.AutoJoinRoomHandler(bot),
// 	)

// 	bot.AddEventHandler(
// 		mxbot.NewLoggerHandler("ada"),
// 	)

// 	bobr := bobrix.NewBobrix(bot, bobrix.WithHealthcheck(bobrix.WithAutoSwitch()))

// 	bobr.SetContractParser(bobrix.DefaultContractParser(bobr.Bot()))

// 	bobr.SetContractParser(bobrix.AutoRequestParser(&bobrix.AutoParserOpts{
// 		Bot:         bot,
// 		ServiceName: "ada",
// 		MethodName:  "generate",
// 		InputName:   "prompt",
// 	}))

// 	bobr.ConnectService(NewADAService("hilltwinssl.singularitynet.io"), func(ctx mxbot.Ctx, r *contracts.MethodResponse, extra any) {

// 		if r.Err != nil {
// 			slog.Error("failed to process request", "error", r.Err)

// 			if err := ctx.TextAnswer("error: " + r.Err.Error()); err != nil {
// 				slog.Error("failed to send message", "error", err)
// 			}

// 			return
// 		}

// 		answer, ok := r.GetString("text")
// 		if !ok {
// 			answer = "I don't know"
// 		}

// 		slog.Debug("got response", "answer", answer)

// 		if err := ctx.TextAnswer(answer); err != nil {
// 			slog.Error("failed to send message", "error", err)
// 		}
// 	})

// 	return bobr, nil
// }
