package bot

import (
	"context"

	domainbot "github.com/tensved/bobrix/mxbot/domain/bot"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
)

var _ domainbot.OlmMachine = (*OlmAdapter)(nil)

type OlmAdapter struct {
	machine *crypto.OlmMachine
}

func NewOlmAdapter(machine *crypto.OlmMachine) *OlmAdapter {
	return &OlmAdapter{machine: machine}
}

func (o *OlmAdapter) DecryptEvent(
	ctx context.Context,
	evt *event.Event,
) (*event.Event, error) {
	return o.machine.DecryptMegolmEvent(ctx, evt)
}
