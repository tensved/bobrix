package mxbot

import (
	"context"
	"maunium.net/go/mautrix/event"
)

func AutoJoinRoomHandler(bot Bot) EventHandler {
	return NewStateMemberHandler(func(ctx Ctx) error {
		evt := ctx.Event()

		if evt.Content.AsMember().Membership == event.MembershipInvite {
			err := bot.JoinRoom(context.TODO(), evt.RoomID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
