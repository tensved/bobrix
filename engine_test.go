package bobrix

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
)

// stubBot implements only what Engine.DisconnectBot touches; any other
// method panics on the nil embedded interface.
type stubBot struct {
	mxbot.Bot
	name     string
	stopped  bool
	closed   bool
	closeErr error
}

func (b *stubBot) Name() string                            { return b.name }
func (b *stubBot) StopListening(ctx context.Context) error { b.stopped = true; return nil }
func (b *stubBot) Close() error                            { b.closed = true; return b.closeErr }

func newStubBobrix(name string, svcIDs ...uuid.UUID) (*Bobrix, *stubBot) {
	bot := &stubBot{name: name}
	bx := NewBobrix(bot)
	for _, id := range svcIDs {
		bx.ConnectService(&contracts.Service{ID: id, Name: id.String()}, nil)
	}
	return bx, bot
}

func connectToEngine(e *Engine, bx *Bobrix) {
	e.ConnectBot(bx)
	for _, svc := range bx.Services() {
		e.ConnectService(svc)
	}
}

func TestEngineDisconnectBot_RemovesBotAndItsServices(t *testing.T) {
	shared := uuid.New()
	own := uuid.New()

	e := NewEngine()
	target, targetBot := newStubBobrix("target", own, shared)
	multibot, multibotBot := newStubBobrix("multibot", shared)
	connectToEngine(e, target)
	connectToEngine(e, multibot)

	bots := e.Bots() // a slice taken before removal must stay intact

	ok, err := e.DisconnectBot(context.Background(), "target")
	if err != nil || !ok {
		t.Fatalf("DisconnectBot() = %v, %v; want true, nil", ok, err)
	}

	if !targetBot.stopped || !targetBot.closed {
		t.Errorf("target bot not stopped: stopped=%v closed=%v", targetBot.stopped, targetBot.closed)
	}
	if multibotBot.stopped {
		t.Error("multibot was stopped")
	}
	if e.GetBot("target") != nil {
		t.Error("GetBot(target) still returns the bot")
	}
	if e.GetBot("multibot") != multibot {
		t.Error("multibot was removed")
	}
	if len(bots) != 2 || bots[0] != target || bots[1] != multibot {
		t.Error("previously returned Bots() slice was mutated")
	}

	if e.GetService(own) != nil {
		t.Error("removed bot's own service still in engine")
	}
	// The shared service must survive through the multibot's entry.
	svc := e.GetService(shared)
	if svc == nil {
		t.Fatal("shared service removed from engine")
	}
	if got, _ := multibot.GetServiceByID(shared); svc != got {
		t.Error("shared service left in engine is not the multibot's")
	}
}

func TestEngineDisconnectBot_UnknownName(t *testing.T) {
	e := NewEngine()
	bx, bot := newStubBobrix("present")
	connectToEngine(e, bx)

	ok, err := e.DisconnectBot(context.Background(), "absent")
	if ok || err != nil {
		t.Fatalf("DisconnectBot() = %v, %v; want false, nil", ok, err)
	}
	if bot.stopped || e.GetBot("present") == nil {
		t.Error("unrelated bot touched")
	}
}

func TestEngineDisconnectBot_RemovesEvenIfStopFails(t *testing.T) {
	e := NewEngine()
	bx, bot := newStubBobrix("broken")
	bot.closeErr = errors.New("close failed")
	connectToEngine(e, bx)

	ok, err := e.DisconnectBot(context.Background(), "broken")
	if !ok || err == nil {
		t.Fatalf("DisconnectBot() = %v, %v; want true, error", ok, err)
	}
	if e.GetBot("broken") != nil {
		t.Error("bot left in engine after failed stop")
	}
}
