package dispatcher // ok

import (
	"context"

	"github.com/tensved/bobrix/mxbot/application/ctx"
	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/filters"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

type Dispatcher struct {
	bot      bot.FullBot
	factory  ctx.Factory
	handlers []handlers.EventHandler
	filters  []filters.Filter
}

// ???? here or not here
func (d *Dispatcher) Dispatch(ctx context.Context, evt *event.Event) error {
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
