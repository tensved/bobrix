package mxbot

import (
	"context"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"slices"
	"strings"
	"time"
)

// Filter - message filter
// return true if message should be processed
// return false if message should be ignored
type Filter func(evt *event.Event) bool

// FilterNotMe - filter for messages from other users
// (ignores messages from the bot itself)
func FilterNotMe(matrixClient *mautrix.Client) Filter {
	return func(evt *event.Event) bool {
		return evt.Sender != matrixClient.UserID
	}
}

type FilterAfterStartOptions struct {
	StartTime      time.Time // start time. Messages sent before this time will be ignored. Default: time.Now()
	ProcessInvites bool      // if true, filter will skip invite filters to chat.
}

// FilterAfterStart - filter for messages after start time
// (ignores messages that were sent before start time)
func FilterAfterStart(bot Bot, opts ...FilterAfterStartOptions) Filter {
	params := FilterAfterStartOptions{
		StartTime:      time.Now(),
		ProcessInvites: false,
	}

	filter := FilterInviteMe(bot)

	if len(opts) > 0 {
		params = opts[0]
	}
	return func(evt *event.Event) bool {
		if params.ProcessInvites && filter(evt) {
			return FilterNotInRoom(bot)(evt)
		}

		eventTime := time.UnixMilli(evt.Timestamp)
		return eventTime.After(params.StartTime)
	}
}

func FilterNotInRoom(bot Bot) Filter {
	return func(evt *event.Event) bool {

		_, err := bot.Client().JoinedMembers(context.Background(), evt.RoomID)

		return err != nil
	}
}

// FilterCommand - filter for command messages
// (check if message starts with command prefix and name)
func FilterCommand(command *Command) Filter {
	return func(evt *event.Event) bool {

		if evt.Type != event.EventMessage {
			return false
		}

		msg := evt.Content.AsMessage().Body
		wordsInMessage := strings.Split(msg, " ")

		if len(wordsInMessage) < 1 {
			return false
		}

		commandPrefix := command.Prefix + command.Name
		return strings.EqualFold(wordsInMessage[0], commandPrefix)
	}
}

func FilterPrivateRoom(cl *mautrix.Client) Filter {
	return func(evt *event.Event) bool {

		resp, err := cl.JoinedMembers(context.Background(), evt.RoomID)
		if err != nil {
			slog.Error("cannot get room joined members", "err", err, "event_id", evt.ID)
			return false
		}

		return len(resp.Joined) == 2
	}
}

func FilterTagMe(userID id.UserID) Filter {

	return func(evt *event.Event) bool {

		mentions := evt.Content.AsMessage().Mentions

		if mentions == nil {
			return false
		}

		for _, mention := range mentions.UserIDs {
			if mention == userID {
				return true
			}
		}

		return false
	}
}

func FilterTageMeOrPrivate(cl *mautrix.Client) Filter {

	return func(evt *event.Event) bool {
		return FilterTagMe(cl.UserID)(evt) || FilterPrivateRoom(cl)(evt)
	}
}

// FilterEventTypes - filter for specific event types
// check if event type is in the list
// return true if event type is in the list
func FilterEventTypes(eventTypes ...event.Type) Filter {
	return func(evt *event.Event) bool {
		return slices.Contains(eventTypes, evt.Type)
	}
}

// FilterEventMessage - filter for event messages
// check if event type is event.EventMessage
// return true if event type is event.EventMessage
func FilterEventMessage() Filter {
	return FilterEventTypes(event.EventMessage)
}

// FilterMessageTypes - filter for specific message types
// check if message type is in the list of message types (event.MessageType)
// also check if event type is event.EventMessage
// return true if message type is in the list
func FilterMessageTypes(msgTypes ...event.MessageType) Filter {
	return func(evt *event.Event) bool {
		return FilterEventMessage()(evt) &&
			slices.Contains(msgTypes, evt.Content.AsMessage().MsgType)
	}
}

// FilterMessageText - filter for text messages
// check if message type is event.MsgText
// return true if message type is event.MsgText
func FilterMessageText() Filter {
	return FilterMessageTypes(event.MsgText)
}

// FilterMessageAudio - filter for audio messages
// check if message type is event.MsgAudio
// return true if message type is event.MsgAudio
func FilterMessageAudio() Filter {
	return FilterMessageTypes(event.MsgAudio)
}

// FilterMessageImage - filter for image messages
// check if message type is event.MsgImage
// return true if message type is event.MsgImage
func FilterMessageImage() Filter {
	return FilterMessageTypes(event.MsgImage)
}

// FilterMessageVideo - filter for video messages
// check if message type is event.MsgVideo
// return true if message type is event.MsgVideo
func FilterMessageVideo() Filter {
	return FilterMessageTypes(event.MsgVideo)
}

// FilterMessageFile - filter for file messages
// check if message type is event.MsgFile
// return true if message type is event.MsgFile
func FilterMessageFile() Filter {
	return FilterMessageTypes(event.MsgFile)
}

func FilterMembershipEvent(membership event.Membership) Filter {
	return func(evt *event.Event) bool {
		return evt.Content.AsMember().Membership == membership
	}
}

func FilterMembershipInvite() Filter {
	return FilterMembershipEvent(event.MembershipInvite)
}

func FilterInviteMe(bot Bot) Filter {
	inviteEventFilter := FilterMembershipInvite()
	return func(evt *event.Event) bool {
		if !inviteEventFilter(evt) {
			return false
		}

		stateKey := evt.StateKey

		if stateKey == nil {
			return false
		}

		return *stateKey == bot.FullName()
	}
}
