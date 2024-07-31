package bobrix

import (
	"context"
	"errors"
	"fmt"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
	"log/slog"
	"maunium.net/go/mautrix/event"
)

type ServiceHandler func(ctx mxbot.Ctx, r *contracts.MethodResponse)

// BobrixService - service for bot
// Binds AI service and adds handler for processing
type BobrixService struct {
	Service *contracts.Service
	Handler ServiceHandler
}

// Bobrix - AI-bot structure
// It is a linking structure between two libs: it manages the bot and AI services
type Bobrix struct {
	bot      mxbot.Bot
	services []*BobrixService
}

// NewBobrix - Bobrix constructor
func NewBobrix(mxBot mxbot.Bot) *Bobrix {
	return &Bobrix{
		bot:      mxBot,
		services: make([]*BobrixService, 0),
	}
}

func (m *Bobrix) Run(ctx context.Context) error {
	return m.bot.StartListening(ctx)
}

func (m *Bobrix) Stop(ctx context.Context) error {
	return m.bot.StopListening(ctx)
}

// ConnectService - add service to the bot
// It is used for adding AI services
// It adds handler for processing the events of the service
func (m *Bobrix) ConnectService(service *contracts.Service, handler func(ctx mxbot.Ctx, r *contracts.MethodResponse)) {
	m.services = append(m.services, &BobrixService{
		Service: service,
		Handler: handler,
	})
}

// Use - add handler to the bot
// It is used for adding event handlers (like middlewares or any other handler)
func (m *Bobrix) Use(handler mxbot.EventHandler) {
	m.bot.AddEventHandler(handler)
}

func (m *Bobrix) GetService(name string) (*BobrixService, bool) {
	for _, botService := range m.services {
		if botService.Service.Name == name {
			return botService, true
		}
	}
	return nil, false
}

type AIRequest struct {
	ServiceName string
	MethodName  string
	InputParams map[string]any
}

type AIHandle func(evt *event.Event) *AIRequest

func (m *Bobrix) SetContractParser(parser func(evt *event.Event) *AIRequest) {

	m.Use(mxbot.NewEventHandler(
		event.EventMessage,
		func(ctx mxbot.Ctx) error {
			request := parser(ctx.Event())

			if request == nil {
				return nil
			}

			botService, ok := m.GetService(request.ServiceName)
			if !ok {
				slog.Error("service not found", "service", request.ServiceName)
				if err := ctx.Answer(fmt.Sprintf("service \"%s\" not found", request.ServiceName)); err != nil {
					return err
				}

				return nil
			}

			service := botService.Service
			handler := botService.Handler

			response, err := service.CallMethod(request.MethodName, request.InputParams)
			if err != nil {
				switch {
				case errors.Is(err, contracts.ErrMethodNotFound):
					if err := ctx.Answer(fmt.Sprintf("method \"%s\" not found", request.MethodName)); err != nil {
						return err
					}

				default:
					if err := ctx.Answer(fmt.Sprintf("error: %s", err)); err != nil {
						return err
					}
				}

				return nil
			}

			handler(ctx, response)

			return nil

		}),
	)
}
