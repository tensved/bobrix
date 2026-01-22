package handlers

import (
	applfilters "github.com/tensved/bobrix/mxbot/application/filters"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domctx "github.com/tensved/bobrix/mxbot/domain/ctx"
	domhandlers "github.com/tensved/bobrix/mxbot/domain/handlers"
)

type JoinRoomParams struct {
	PreJoinHook   func(domctx.Ctx) error
	AfterJoinHook func(domctx.Ctx) error
}

// AutoJoinRoomHandler - join the room on invite automatically
// You can pass JoinRoomParams to modify the behavior of the handler
// Use PreJoinHook to modify the behavior before joining the room
// If PreJoinHook returns an error, the join is aborted
// Use AfterJoinHook to modify the behavior after joining the room
func AutoJoinRoomHandler(
	room dombot.BotRoomActions,
	info dombot.BotInfo,
	params ...JoinRoomParams,
) domhandlers.EventHandler {
	return domhandlers.NewStateMemberHandler(func(ctx domctx.Ctx) error {
		evt := ctx.Event()

		var p JoinRoomParams
		if len(params) > 0 {
			p = params[0]
		}

		if p.PreJoinHook != nil {
			if err := p.PreJoinHook(ctx); err != nil {
				return err
			}
		}

		if err := room.JoinRoom(ctx.Context(), evt.RoomID); err != nil {
			return err
		}

		if p.AfterJoinHook != nil {
			if err := p.AfterJoinHook(ctx); err != nil {
				return err
			}
		}

		return nil
	}, applfilters.FilterInviteMe(info))
}
