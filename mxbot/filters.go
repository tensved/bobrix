package mxbot

import (
	"github.com/tensved/bobrix/mxbot/application/filters"
	dombot "github.com/tensved/bobrix/mxbot/domain/bot"
	domfilters "github.com/tensved/bobrix/mxbot/domain/filters"
)

// ---- message type ----

func FilterMessageText() domfilters.Filter {
	return filters.FilterMessageText()
}

func FilterMessageAudio() domfilters.Filter {
	return filters.FilterMessageAudio()
}

// ---- bot related ----

func FilterTagMe(bot dombot.BotInfo) domfilters.Filter {
	return filters.FilterTagMe(bot)
}

func FilterPrivateRoom(bot dombot.BotRoomActions) domfilters.Filter {
	return filters.FilterPrivateRoom(bot)
}

func FilterTagMeOrPrivate(info dombot.BotInfo, room dombot.BotRoomActions) domfilters.Filter {
	return filters.FilterTagMeOrPrivate(info, room)
}
