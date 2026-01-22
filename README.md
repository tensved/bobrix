[![Go Reference](https://pkg.go.dev/badge/github.com/tensved/bobrix.svg)](https://pkg.go.dev/github.com/tensved/bobrix)
[![Go Report Card](https://goreportcard.com/badge/github.com/tensved/bobrix)](https://goreportcard.com/report/github.com/tensved/bobrix)

# BOBRIX (Bot Bridge Matrix) -  Service Bridge for Matrix Bots

### Overview

BOBRIX (Bot Bridge Matrix) is a Go library designed to facilitate interaction
with various services through a Matrix client in the form of a bot. BoBRIX abstracts the complexity of integrating
multiple services by providing a unified set of interaction contracts and a convenient layer
on top of the Matrix client, enabling seamless service integration.

### Features
- **Interaction Contracts**: Define and manage interaction protocols with various services.
- **Matrix Client Bot**: A high-level abstraction over the Matrix client to simplify bot development.
- **Service Integration**: Combine interaction contracts and the bot framework to interact with multiple
services directly from the Matrix client.

## Getting started

### Prerequisites

BoBRIX requires [Go](https://go.dev/) version [1.21](https://go.dev/doc/devel/release#go1.21.0) or above.

### BoBRIX Installation

To install BoBRIX, use the following command:
```sh
go get -u github.com/tensved/bobrix
```


### Usage

A basic example of creating bobrix Engine:
```go
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
			HomeServerURL: os.Getenv("MX_BOT_HOMESERVER_URL"),
			PickleKey:     []byte(os.Getenv("PICKLE_KEY")),
		},
		TypingTimeout: 3 * time.Second,
		SyncTimeout:   5 * time.Second,
	}

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
```

A basic creating matrix bot:

```go
package ada

import (
	"log/slog"

	"github.com/tensved/bobrix"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
)

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
```

Basic example of Service Description:

```go
package ada

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tensved/bobrix/contracts"
	"log/slog"
	"net/url"
)

func NewADAService(adaHost string) *contracts.Service {

	return &contracts.Service{
		Name:        "ada",
		Description: "Ada is an AI language model trained by OpenAI.",
		Methods: map[string]*contracts.Method{
			"generate": {
				Name:        "generate",
				Description: "Generate text using Ada's language model.",
				Inputs: []contracts.Input{
					{
						Name:        "prompt",
						Description: "The prompt to generate text from.",
						Type:        "text",
					},
					{
						Name:         "response_type",
						Type:         "text",
						DefaultValue: "text",
					},
					{
						Name: "audio",
						Type: "audio",
					},
				},
				Outputs: []contracts.Output{
					{
						Name:        "text",
						Description: "The generated text.",
					},
				},
				Handler: NewADAHandler(adaHost),
			},
		},
		Pinger: contracts.NewWSPinger(
			contracts.WSOptions{Host: adaHost, Path: "/", Schema: "wss"},
		),
	}
}
```
*For a more detailed description see [Service Example](examples/ada/service.go)*

```
┌────────────────────────────┐
│ mautrix.DefaultSyncer      │
│ OnEvent(ctx, evt)          │
└────────────┬───────────────┘
             │
             ▼
┌────────────────────────────┐
│ events.Service             │
│ HandleMatrixEvent          │
└────────────┬───────────────┘
             │
     ┌───────┴───────────────────────────┐
     │                                    │
     ▼                                    ▼
┌───────────────┐               ┌──────────────────┐
│ to-device evt │               │ room / message   │
│ m.room_key    │               │ m.room.message   │
└───────┬───────┘               └────────┬─────────┘
        │                                │
        ▼                                ▼
┌──────────────────┐           ┌────────────────────┐
│ crypto.Handle    │           │ crypto.Decrypt     │
│ ToDeviceEvent    │           │ (Megolm)           │
└──────────────────┘           └────────┬───────────┘
                                         │
                           decrypt error │ success
                                         │
                    ┌────────────────────┴────────────┐
                    ▼                                 ▼
        auto RequestRoomKey                decrypted event
                                                  │
                                                  ▼
                                    ┌────────────────────────┐
                                    │ transport filters      │
                                    └────────┬───────────────┘
                                             │
                                             ▼
                                    ┌────────────────────────┐
                                    │ dispatcher / handlers  │
                                    │ ctx.Ctx                │
                                    └────────────────────────┘

```