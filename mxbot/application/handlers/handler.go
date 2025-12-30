package handlers

import "github.com/tensved/bobrix/mxbot/domain/ctx"

type EventHandler interface {
	Handle(ctx.Ctx) error
}
