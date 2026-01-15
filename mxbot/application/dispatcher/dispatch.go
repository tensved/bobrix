package dispatcher

import (
	"context"
	"fmt"

	// "github.com/tensved/bobrix/mxbot/application/ctx"
	// "github.com/tensved/bobrix/mxbot/domain/bot"
	// "github.com/tensved/bobrix/mxbot/domain/filters"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

// func (d *Dispatcher) Dispatch(ctx context.Context, evt *event.Event) error {
// 	if evt == nil {
// 		return nil
// 	}

// 	// 1️⃣ decrypt if needed
// 	if d.bot.IsEncrypted(evt) {
// 		decrypted, err := d.bot.DecryptEvent(ctx, evt)
// 		if err != nil {
// 			return err
// 		}
// 		evt = decrypted
// 	}

// 	// 2️⃣ global filters (before creating Ctx)
// 	for _, f := range d.filters {
// 		if !f(evt) {
// 			return nil
// 		}
// 	}

// 	// 3️⃣ create context
// 	c, err := d.factory.New(ctx, evt)
// 	if err != nil {
// 		return err
// 	}

// 	// 4️⃣ run handlers
// 	for _, h := range d.handlers {
// 		if c.IsHandled() {
// 			return nil
// 		}

// 		if h.EventType() != evt.Type {
// 			continue
// 		}

// 		if err := h.Handle(c); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (d *Dispatcher) Dispatch(ctx context.Context, evt *event.Event) error {
// 	c, err := d.factory.New(ctx, evt)
// 	if err != nil {
// 		return err
// 	}

// 	for _, h := range d.handlers {
// 		if err := h.Handle(c); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

var _ bot.EventSink = (*Dispatcher)(nil)

func (d *Dispatcher) HandleMatrixEvent(ctx context.Context, evt *event.Event) error {
	if d.bot == nil {
		return fmt.Errorf("dispatcher: bot is not set")
	}

	c, err := d.factory.New(ctx, evt)
	if err != nil {
		return err
	}

	for _, h := range d.handlers {
		if err := h.Handle(c); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) SetBot(b bot.FullBot) {
	d.bot = b
}

var _ bot.EventDispatcher = (*Dispatcher)(nil)

func (d *Dispatcher) AddEventHandler(h handlers.EventHandler) {
	d.handlers = append(d.handlers, h)
}
