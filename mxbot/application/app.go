// application/app.go
package application

import "github.com/tensved/bobrix/mxbot/domain/bot"

type App struct {
	bot bot.FullBot
}

func New(bot bot.FullBot) *App {
	return &App{bot: bot}
}
