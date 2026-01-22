package mxbot

import (
	"log/slog"

	"maunium.net/go/mautrix/event"

	applhandlers "github.com/tensved/bobrix/mxbot/application/handlers"

	"github.com/tensved/bobrix/mxbot/domain/ctx"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
)

func NewEventHandler(
	eventType event.Type,
	handler func(Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewEventHandler(eventType, handler, filters...)
}

func NewStateMemberHandler(
	handler func(ctx.Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewStateMemberHandler(handler, filters...)
}

func NewMessageHandler(
	handler func(Ctx) error,
	filters ...filters.Filter,
) EventHandler {
	return handlers.NewMessageHandler(handler, filters...)
}

func AutoJoinRoomHandler(bot Bot, params ...applhandlers.JoinRoomParams) EventHandler {
	return applhandlers.AutoJoinRoomHandler(bot, bot, params...)
}

func NewLoggerHandler(name string, log ...*slog.Logger) EventHandler {
	return applhandlers.NewLoggerHandler(name, log...)
}

type JoinRoomParams = applhandlers.JoinRoomParams
