package main

import (
	"context"
	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/examples/ada"
	"github.com/tensved/bobrix/mxbot"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	engine := bobrix.NewEngine()

	botCredentials := &mxbot.BotCredentials{
		Username:      os.Getenv("MX_BOT_USERNAME"),
		Password:      os.Getenv("MX_BOT_PASSWORD"),
		HomeServerURL: os.Getenv("MX_BOT_HOMESERVER_URL"),
	}
	adaBot, err := ada.NewAdaBot(botCredentials)
	if err != nil {
		panic(err)
	}

	engine.ConnectBot(adaBot)

	ctx := context.Background()

	if err := engine.Run(ctx); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := engine.Stop(ctx); err != nil {
		slog.Error("failed to stop engine", "error", err)
	}

	slog.Info("service shutdown")
}
