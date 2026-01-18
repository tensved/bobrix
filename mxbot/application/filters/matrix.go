package filters

import (
	"context"
	"log/slog"
	"time"

	db "github.com/tensved/bobrix/mxbot/domain/bot"
	df "github.com/tensved/bobrix/mxbot/domain/filters"
	"maunium.net/go/mautrix/event"
)

type FilterAfterStartOptions struct {
	StartTime      time.Time // start time. Messages sent before this time will be ignored. Default: time.Now()
	ProcessInvites bool      // if true, filter will skip invite filters to chat.
}

// FilterAfterStart - filter for messages after start time
// (ignores messages that were sent before start time)
func FilterAfterStart(
	info db.BotInfo,
	room db.BotRoomActions,
	opts ...FilterAfterStartOptions,
) df.Filter {
	params := FilterAfterStartOptions{
		StartTime:      time.Now(),
		ProcessInvites: false,
	}

	filter := FilterInviteMe(info)

	if len(opts) > 0 {
		params = opts[0]
	}
	return func(evt *event.Event) bool {
		if params.ProcessInvites && filter(evt) {
			return FilterNotInRoom(room)(evt)
		}

		eventTime := time.UnixMilli(evt.Timestamp)
		return eventTime.After(params.StartTime)
	}
}

// FilterTageMeOrPrivate - filter for messages that are tagged or sent in a private room
// return true if message is tagged or sent in a private room
func FilterTagMeOrPrivate(
	info db.BotInfo,
	room db.BotRoomActions,
) df.Filter {
	return FilterAny(
		FilterTagMe(info),
		FilterPrivateRoom(room),
	)
}

// FilterNotInRoom - filter for messages that bot is not in the room
func FilterNotInRoom(r db.BotRoomActions) df.Filter {
	return func(evt *event.Event) bool {
		_, err := r.JoinedMembersCount(context.Background(), evt.RoomID)
		return err != nil
	}
}

// FilterPrivateRoom - filter for private rooms (there are only two people in the room: bot + user)
// return true if room is private
func FilterPrivateRoom(r db.BotRoomActions) df.Filter {
	return func(evt *event.Event) bool {
		count, err := r.JoinedMembersCount(context.Background(), evt.RoomID)
		if err != nil {
			slog.Error("joined members error", "err", err)
			return false
		}
		return count == 2
	}
}

// FilterNotMe - filter for messages from other users
// (ignores messages from the bot itself)
func FilterNotMe(bot db.BotInfo) df.Filter {
	return func(evt *event.Event) bool {
		return evt.Sender != bot.UserID()
	}
}

// FilterInviteMe - filter for invite messages that are sent to the bot
// check if message type is event.MembershipInvite and state key is the bot's full name
func FilterInviteMe(bot db.BotInfo) df.Filter {
	return func(evt *event.Event) bool {
		if evt.Content.AsMember().Membership != event.MembershipInvite {
			return false
		}
		return evt.StateKey != nil && *evt.StateKey == bot.FullName()
	}
}

// FilterTagMe - filter for messages that bot is tagged
// return true if bot is tagged
func FilterTagMe(bot db.BotInfo) df.Filter {
	return func(evt *event.Event) bool {
		msg := evt.Content.AsMessage()
		if msg.Mentions == nil {
			return false
		}
		for _, id := range msg.Mentions.UserIDs {
			if id == bot.UserID() {
				return true
			}
		}
		return false
	}
}

// FilterMembershipEvent - filter for membership events
// check if message type is event.Membership
func FilterMembershipEvent(m event.Membership) df.Filter {
	return func(evt *event.Event) bool {
		return evt.Content.AsMember().Membership == m
	}
}

// FilterMembershipInvite - filter for invite messages
// check if message type is event.MembershipInvite
func FilterMembershipInvite() df.Filter {
	return FilterMembershipEvent(event.MembershipInvite)
}
