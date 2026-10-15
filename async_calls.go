package bobrix

import (
	"runtime/debug"

	"github.com/tensved/bobrix/mxbot"
)

// WithAsyncCalls runs service calls outside the Matrix sync worker, so a slow
// service (e.g. a model thinking for minutes) does not block the processing
// of other events. At most maxConcurrent calls run at once; when all slots are
// busy the worker waits for a free one, which gives back-pressure to the sync
// queue instead of an unbounded pile-up of goroutines.
//
// The event is considered processed as soon as the call is started: errors of
// the call are logged and answered to the user, not returned to the worker.
// maxConcurrent < 1 keeps calls synchronous.
func WithAsyncCalls(maxConcurrent int) BobrixOpts {
	return func(bx *Bobrix) {
		if maxConcurrent < 1 {
			return
		}
		bx.callSlots = make(chan struct{}, maxConcurrent)
	}
}

// dispatchServiceRequest handles a claimed request in place or, with
// WithAsyncCalls, in a separate goroutine.
func (bx *Bobrix) dispatchServiceRequest(ctx mxbot.Ctx, req *ServiceRequest, opt ContractParserOpts) error {
	if bx.callSlots == nil {
		return bx.handleServiceRequest(ctx, req, opt)
	}

	select {
	case bx.callSlots <- struct{}{}:
	case <-ctx.Context().Done():
		return ctx.Context().Err()
	}

	go func() {
		defer func() { <-bx.callSlots }()
		defer func() {
			if r := recover(); r != nil {
				bx.logger.Error("service request panic recovered",
					"panic", r, "stack", string(debug.Stack()),
					"event_id", ctx.Event().ID)
			}
		}()

		if err := bx.handleServiceRequest(ctx, req, opt); err != nil {
			bx.logger.Error("service request failed",
				"err", err,
				"service_id", req.ServiceID,
				"service", req.ServiceName,
				"event_id", ctx.Event().ID)
		}
	}()

	return nil
}
