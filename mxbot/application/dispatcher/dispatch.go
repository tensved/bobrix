package dispatcher

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tensved/bobrix/mxbot/domain/filters"

	"github.com/tensved/bobrix/mxbot/domain/bot"
	"github.com/tensved/bobrix/mxbot/domain/handlers"
	"maunium.net/go/mautrix/event"
)

var _ bot.EventDispatcher = (*Dispatcher)(nil)
var _ bot.EventRouter = (*Dispatcher)(nil)
var _ bot.EventSink = (*Dispatcher)(nil)

func (d *Dispatcher) HandleMatrixEvent(ctx context.Context, evt *event.Event) error {
	slog.Info(
		"DISPATCHER: event",
		"type", evt.Type,
		"content", evt.Content.Raw,
	)

	slog.Info(
		"DISPATCHER Handle",
		"ptr", fmt.Sprintf("%p", d),
		"handlers", len(d.handlers),
	)

	if d.bot == nil {
		return fmt.Errorf("dispatcher: bot is not set")
	}

	slog.Info("!!!!!!!!!!!!!1")

	// --- global filters
	for _, f := range d.filters {
		if !f(evt) {
			return nil
		}
	}

	slog.Info("!!!!!!!!!!!!!2")

	slog.Info(
		"MESSAGE EVENT",
		"roomID", evt.RoomID,
		"type", evt.Type,
	)

	// --- typing ONLY for messages
	var cancelTyping func()
	if evt.Type == event.EventMessage {
		var err error
		cancelTyping, err = d.bot.LoopTyping(ctx, evt.RoomID)
		if err != nil {
			slog.Info("!!!!!!!!!!!!!11")
			d.logger.Warn().Err(err).Msg("failed to start typing")
		} else {
			slog.Info("!!!!!!!!!!!!!12")
			defer cancelTyping()
		}
	}

	// cancelTyping, err := d.bot.LoopTyping(ctx, evt.RoomID)
	// if err != nil {
	// 	slog.Info("!!!!!!!!!!!!!11")
	// 	d.logger.Warn().Err(err).Msg("failed to start typing")
	// } else {
	// 	slog.Info("!!!!!!!!!!!!!12")
	// 	defer cancelTyping()
	// }

	slog.Info("!!!!!!!!!!!!!3")

	eventContext, err := d.factory.New(ctx, evt)
	defer eventContext.SetHandled()
	if err != nil {
		return err
	}

	slog.Info("!!!!!!!!!!!!!4")

	// for _, h := range d.handlers {
	// 	if h.EventType() != evt.Type {
	// 		continue
	// 	}
	// 	if err := h.Handle(eventContext); err != nil {
	// 		return err
	// 	}
	// }

	for _, h := range d.handlers {
		slog.Info(
			"CHECK HANDLER",
			"handler", fmt.Sprintf("%T", h),
			"handlerType", h.EventType(),
			"eventType", evt.Type,
		)

		if h.EventType() != evt.Type {
			continue
		}
		slog.Info("HANDLER MATCHED", "handler", fmt.Sprintf("%T", h))

		// cancelTyping, _ := d.bot.LoopTyping(ctx, evt.RoomID)
		// if cancelTyping != nil {
		// 	defer cancelTyping()
		// }

		// for i, f := range d.filters {
		// 	ok := f(evt)
		// 	slog.Info("GLOBAL FILTER",
		// 		"index", i,
		// 		"eventType", evt.Type,
		// 		"result", ok,
		// 	)
		// 	if !ok {
		// 		return nil
		// 	}
		// }

		if err := h.Handle(eventContext); err != nil {
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

func (d *Dispatcher) EventHandlers() []handlers.EventHandler {
	return d.handlers
}

func (d *Dispatcher) Filters() []filters.Filter {
	return d.filters
}

func (d *Dispatcher) AddFilter(f filters.Filter) {
	d.filters = append(d.filters, f)
}


