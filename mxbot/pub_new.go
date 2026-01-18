package mxbot

import (
	"context"
	"time"

	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/threads"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

//
// ===== BOT FACADE =====
//

// Bot — публичный интерфейс Matrix-бота
// Единственное, что знает внешний мир
type Bot interface {
	StartListening(ctx context.Context) error
	StopListening(ctx context.Context) error

	AddEventHandler(handler EventHandler)

	SetOnlineStatus()
	SetOfflineStatus()
	SetIdleStatus()

	Ping(ctx context.Context) error

	BotInfo
	BotRoomActions
}

//
// ===== CONTEXT =====
//

// Ctx — публичный контекст события
// Реальная реализация скрыта
type Ctx interface {
	Context() context.Context
	Event() *event.Event

	TextAnswer(text string) error
	ErrorAnswer(msg string, code int) error

	Thread() *threads.MessagesThread

	// управление жизненным циклом события
	SetHandled()
	IsHandled() bool
	IsHandledWithUnlocker() (bool, func())

	// доступ к боту (read-only!)
	Bot() BotInfo
}

type BotInfo interface {
	UserID() id.UserID
	FullName() string // get full name with servername (e.g. @username:servername)
	Name() string
}

//
// ===== FILTERS =====
//

// Filter — публичный alias
type Filter = filters.Filter

type BotRoomActions interface {
	JoinRoom(ctx context.Context, roomID id.RoomID) error
	JoinedMembersCount(ctx context.Context, roomID id.RoomID) (int, error)

	GetJoinedRoomsList(ctx context.Context) ([]id.RoomID, error)
	GetMessagesFromRoomByNumber(ctx context.Context, roomID id.RoomID, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error)
	GetMessagesFromRoomByDuration(ctx context.Context, roomID id.RoomID, duration time.Duration, numMessages int, filter *mautrix.FilterPart) ([]*event.Event, error)
}
