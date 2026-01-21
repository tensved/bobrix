package ada

import (
	"log/slog"

	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
)

// go build
// install_name_tool -change build/libolm.dylib.3.2.16 /usr/local/lib/libolm.dylib ./cmd
// ./cmd
func NewAdaBot(cfg *mxbot.Config) (*bobrix.Bobrix, error) {
	bot, err := mxbot.NewMatrixBot(*cfg)
	if err != nil {
		return nil, err
	}

	bot.AddEventHandler(
		mxbot.TextCommand("ping", func(ctx mxbot.Ctx) error {
			return ctx.TextAnswer("pong")
		}),
	)

	bot.AddEventHandler(mxbot.AutoJoinRoomHandler(bot))
	bot.AddEventHandler(mxbot.NewLoggerHandler("ada"))

	bx := bobrix.NewBobrix(
		bot,
		bobrix.WithHealthcheck(bobrix.WithAutoSwitch()),
	)

	bx.SetContractParser(
		bobrix.AutoRequestParser(&bobrix.AutoParserOpts{
			Bot:         bot,
			ServiceName: "ada",
			MethodName:  "generate",
			InputName:   "prompt",
		}),
	)

	bx.SetContractParser(
		bobrix.DefaultContractParser(bot),
	)

	bx.ConnectService(
		NewADAService("hilltwinssl.singularitynet.io"),
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
