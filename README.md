# BoBRIX (Bot Bridge Matrix) -  Service Bridge for Matrix Bots

### Overview

BoBRIX (Bot Bridge Matrix) is a Go library designed to facilitate interaction
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

	botCredentials := mxbot.BotCredentials{
		Username:      os.Getenv("MX_BOT_USERNAME"),
		Password:      os.Getenv("MX_BOT_PASSWORD"),
		HomeServerURL: os.Getenv("MX_BOT_HOMESERVER_URL"),
	}
	adaBot, err := ada.NewAdaBot(botCredentials.Username, botCredentials.Password, botCredentials.HomeServerURL)
	if err != nil {
		panic(err)
	}

	engine.ConnectBot(adaBot)
}

```

A basic creating matrix bot:

```go
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
				Inputs: []*contracts.Input{
					{
						Name:        "prompt",
						Description: "The prompt to generate text from.",
						IsRequired:  true,
					},
				},
				Outputs: []*contracts.Output{
					{
						Name:        "text",
						Description: "The generated text.",
					},
				},
				Handler: NewADAHandler(adaHost),
			},
		},
	}
}
```
*For a more detailed description see [Service Example](examples/ada/service.go)*

