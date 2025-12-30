package handlers

type JoinRoomParams struct {
	PreJoinHook   func(ctx Ctx) error
	AfterJoinHook func(ctx Ctx) error
}

// AutoJoinRoomHandler - join the room on invite automatically
// You can pass JoinRoomParams to modify the behavior of the handler
// Use PreJoinHook to modify the behavior before joining the room
// If PreJoinHook returns an error, the join is aborted
// Use AfterJoinHook to modify the behavior after joining the room
func AutoJoinRoomHandler(bot BotJoiner, params ...JoinRoomParams) EventHandler {
	return NewStateMemberHandler(func(ctx Ctx) error {
		evt := ctx.Event()

		p := JoinRoomParams{}
		if len(params) > 0 {
			p = params[0]
		}

		if p.PreJoinHook != nil {
			if err := p.PreJoinHook(ctx); err != nil {
				return err
			}
		}

		if err := bot.JoinRoom(ctx.Context(), evt.RoomID); err != nil {
			return err
		}

		if p.AfterJoinHook != nil {
			if err := p.AfterJoinHook(ctx); err != nil {
				return err
			}
		}

		return nil
	}, FilterInviteMe(bot))
}
