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
// Binds service and adds handler for processing
type BobrixService struct {
	Service *contracts.Service
	Handler ServiceHandler

	IsOnline bool
}

// Bobrix - bot structure
// It is a connection structure between two components: it manages the bot and service contracts
type Bobrix struct {
	name          string
	bot           mxbot.Bot
	services      []*BobrixService
	Healthchecker Healthcheck

	logger *slog.Logger
}

type BobrixOpts func(*Bobrix)

func WithHealthcheck(healthCheckOpts ...HealthcheckOption) BobrixOpts {
	return func(bx *Bobrix) {

		healthcheck := NewHealthcheck(bx)

		for _, opt := range healthCheckOpts {
			opt(healthcheck)
		}

		bx.Healthchecker = healthcheck
	}
}

// NewBobrix - Bobrix constructor
func NewBobrix(mxBot mxbot.Bot, opts ...BobrixOpts) *Bobrix {
	bx := &Bobrix{
		name:     mxBot.Name(),
		bot:      mxBot,
		services: make([]*BobrixService, 0),
	}

	bx.logger = slog.Default().With("name", bx.name)

	for _, opt := range opts {
		opt(bx)
	}

	return bx
}

func (bx *Bobrix) Name() string {
	return bx.name
}

func (bx *Bobrix) Run(ctx context.Context) error {
	return bx.bot.StartListening(ctx)
}

func (bx *Bobrix) Stop(ctx context.Context) error {
	return bx.bot.StopListening(ctx)
}

// ConnectService - add service to the bot
// It is used for adding services
// It adds handler for processing the events of the service
func (bx *Bobrix) ConnectService(service *contracts.Service, handler func(ctx mxbot.Ctx, r *contracts.MethodResponse)) {
	bx.services = append(bx.services, &BobrixService{
		Service:  service,
		Handler:  handler,
		IsOnline: true,
	})
}

// Use - add handler to the bot
// It is used for adding event handlers (like middlewares or any other handler)
func (bx *Bobrix) Use(handler mxbot.EventHandler) {
	bx.bot.AddEventHandler(handler)
}

func (bx *Bobrix) GetService(name string) (*BobrixService, bool) {
	for _, botService := range bx.services {
		if botService.Service.Name == name {
			return botService, true
		}
	}
	return nil, false
}

func (bx *Bobrix) Services() []*BobrixService {
	return bx.services
}

func (bx *Bobrix) Bot() mxbot.Bot {
	return bx.bot
}

type ServiceRequest struct {
	ServiceName string
	MethodName  string
	InputParams map[string]any
}

type ServiceHandle func(evt *event.Event) *ServiceRequest

func (bx *Bobrix) SetContractParser(parser func(evt *event.Event) *ServiceRequest) {

	bx.Use(
		mxbot.NewEventHandler(
			event.EventMessage,
			func(ctx mxbot.Ctx) error {

				request := parser(ctx.Event())

				// if request is nil, it means that the event does not match the contract
				// and the event should be ignored
				// or the service is not found
				if request == nil || request.ServiceName == "" {
					return nil
				}

				// not handle if it marked
				if !ctx.CheckAndSetHandled() {
					return nil
				}

				botService, ok := bx.GetService(request.ServiceName)
				if !ok {
					bx.logger.Error("service not found", "service", request.ServiceName)
					if err := ctx.TextAnswer(
						fmt.Sprintf(
							"service \"%s\" not found",
							request.ServiceName,
						),
					); err != nil {
						return err
					}

					return nil
				}

				service := botService.Service
				handler := botService.Handler

				// if service is not online, send an error and return
				// IsOnline is true by default. But it can be changed with Healthcheck with WithAutoSwitch() option
				if !botService.IsOnline {
					handler(ctx, &contracts.MethodResponse{
						Err: fmt.Errorf("service \"%s\" is offline", request.ServiceName),
					})
					return nil
				}

				response, err := service.CallMethod(request.MethodName, request.InputParams)
				if err != nil {
					switch {
					case errors.Is(err, contracts.ErrMethodNotFound):
						if err := ctx.TextAnswer(fmt.Sprintf("method \"%s\" not found", request.MethodName)); err != nil {
							return err
						}

					default:
						if err := ctx.TextAnswer(fmt.Sprintf("error: %s", err)); err != nil {
							return err
						}
					}

					return nil
				}

				handler(ctx, response)

				return nil

			},
		),
	)
}
