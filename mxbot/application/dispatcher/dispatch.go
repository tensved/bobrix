package dispatcher

import (
	"context"
	"fmt"

	"maunium.net/go/mautrix/event"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
)

var _ bot.EventDispatcher = (*Dispatcher)(nil)
var _ bot.EventSink = (*Dispatcher)(nil)

func (d *Dispatcher) HandleMatrixEvent(ctx context.Context, evt *event.Event) error {
	if d.bot == nil {
		return fmt.Errorf("dispatcher: bot is not set")
	}

	// --- global filters
	for _, f := range d.filters {
		if !f(evt) {
			return nil
		}
	}

	// --- typing ONLY for messages
	var cancelTyping func()
	if evt.Type == event.EventMessage {
		var err error
		cancelTyping, err = d.bot.LoopTyping(ctx, evt.RoomID)
		if err != nil {
			d.logger.Warn().Err(err).Msg("failed to start typing")
		} else {
			defer cancelTyping()
		}
	}

	eventContext, err := d.factory.New(ctx, evt)
	defer eventContext.SetHandled()
	if err != nil {
		return err
	}

	for _, h := range d.handlers {
		if h.EventType() != evt.Type {
			continue
		}
		if err := h.Handle(eventContext); err != nil {
			return err
		}
	}

	return nil
}

func (d *Dispatcher) SetBot(b bot.FullBot) {
	d.bot = b
}

func (d *Dispatcher) AddEventHandler(h handlers.EventHandler) {
	d.handlers = append(d.handlers, h)
}

func (d *Dispatcher) EventHandlers() []handlers.EventHandler {
	return d.handlers
}

func (d *Dispatcher) Filters() []filters.Filter {
	return d.filters
}

func (d *Dispatcher) AddFilter(f filters.Filter) {
	d.filters = append(d.filters, f)
}
