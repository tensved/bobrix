package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/examples/ada"
	"github.com/tensved/bobrix/mxbot"
)

func main() {
	engine := bobrix.NewEngine()

	cfg := &mxbot.Config{
		Credentials: &mxbot.BotCredentials{
			Username:      os.Getenv("MX_BOT_USERNAME"),
			Password:      os.Getenv("MX_BOT_PASSWORD"),
			HomeServerURL: "http://localhost:8008",
			PickleKey:     []byte(os.Getenv("PICKLE_KEY")),
		},
		TypingTimeout: 3 * time.Second,
		SyncTimeout:   5 * time.Second,
	}

	// cfg := &mxbot.Config{
	// 	Credentials: &mxbot.BotCredentials{
	// 		Username:      "MX_BOT_USERNAME",
	// 		Password:      "MX_BOT_PASSWORD",
	// 		HomeServerURL: "http://localhost:8008",
	// 		PickleKey:     []byte("V+NSQ5oG2GRdDyTXZKA3dGpgoGXJRL+elIiVTo/9dDI="),
	// 	},
	// TypingTimeout: 3 * time.Second,
	// SyncTimeout:   5 * time.Second,
	// }

	adaBot, err := ada.NewAdaBot(cfg)
	if err != nil {
		panic(err)
	}

	engine.ConnectBot(adaBot)

	ctx := context.Background()

	go func() {
		if err := engine.Run(ctx); err != nil {
			panic(err)
		}
	}()

	go func() {
		ada := engine.Bots()[0]

		slog.Info("starting sync")

		sub := ada.Healthchecker.Subscribe()

		for data := range sub.Sync() {
			slog.Debug("healthcheck", "data", data)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := engine.Stop(ctx); err != nil {
		slog.Error("failed to stop engine", "error", err)
	}

	slog.Info("service shutdown")
}
