package mxbot

import (
	applfilters "github.com/tensved/bobrix/mxbot/application/filters"

	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"
)

// ---- message type ----

func FilterMessageText() domfilters.Filter {
	return applfilters.FilterMessageText()
}

func FilterMessageAudio() domfilters.Filter {
	return applfilters.FilterMessageAudio()
}

func FilterEventMessage() domfilters.Filter {
	return applfilters.FilterEventMessage()
}

// ---- bot related ----

func FilterTagMe(bot dombot.BotInfo) domfilters.Filter {
	return applfilters.FilterTagMe(bot)
}

func FilterPrivateRoom(bot dombot.BotRoomActions) domfilters.Filter {
	return applfilters.FilterPrivateRoom(bot)
}

func FilterTagMeOrPrivate(b Bot) domfilters.Filter {
	return applfilters.FilterTagMeOrPrivate(b, b)
}

func FilterNotMe(bot dombot.BotInfo) domfilters.Filter {
	return applfilters.FilterNotMe(bot)
}
