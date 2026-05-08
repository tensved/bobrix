package bobrix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/tensved/bobrix/contracts"
	"github.com/tensved/bobrix/mxbot"
	"maunium.net/go/mautrix/event"
)

type ServiceHandler func(ctx mxbot.Ctx, r *contracts.MethodResponse, extra any)

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
	name string
	bot  mxbot.Bot

	servicesByID     map[uuid.UUID]*BobrixService
	serviceIDsByName map[string]uuid.UUID // основной алиас = текущее имя

	Healthchecker Healthcheck
	logger        *slog.Logger
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
		name:             mxBot.Name(),
		bot:              mxBot,
		servicesByID:     make(map[uuid.UUID]*BobrixService),
		serviceIDsByName: make(map[string]uuid.UUID),
		logger:           slog.Default().With("name", mxBot.Name()),
	}

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
func (bx *Bobrix) ConnectService(service *contracts.Service, handler ServiceHandler) uuid.UUID {
	if service == nil {
		bx.logger.Error("ConnectService: nil service")
		return uuid.Nil
	}
	if service.ID == uuid.Nil {
		bx.logger.Warn("ConnectService: service has nil ID, generating", "name", service.Name)
		service.ID = uuid.New()
	}

	bx.logger.Info("ConnectService",
		"service_id", service.ID.String(),
		"service_name", service.Name,
	)

	// if service.ID == uuid.Nil {
	// 	service.ID = uuid.New()
	// }

	bs := &BobrixService{
		Service:  service,
		Handler:  handler,
		IsOnline: true,
	}

	bx.servicesByID[service.ID] = bs
	bx.serviceIDsByName[strings.ToLower(service.Name)] = service.ID

	return service.ID
}

// Use - add handler to the bot
// It is used for adding event handlers (like middlewares or any other handler)
func (bx *Bobrix) Use(handler mxbot.EventHandler) {
	bx.bot.AddEventHandler(handler)
}

// GetService - return service by name. If the service is not found, it returns nil
// It is case-insensitive. All services are stored in lowercase
func (bx *Bobrix) GetServiceByID(id uuid.UUID) (*BobrixService, bool) {
	svc, ok := bx.servicesByID[id]
	return svc, ok
}

func (bx *Bobrix) GetServiceByName(name string) (*BobrixService, bool) {
	id, ok := bx.serviceIDsByName[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	return bx.GetServiceByID(id)
}

func (bx *Bobrix) Services() []*BobrixService {
	out := make([]*BobrixService, 0, len(bx.servicesByID))
	for _, svc := range bx.servicesByID {
		out = append(out, svc)
	}
	return out
}

func (bx *Bobrix) Bot() mxbot.Bot {
	return bx.bot
}

type ServiceRequest struct {
	ServiceName string         `json:"service"`
	ServiceID   string         `json:"service_id"`
	MethodName  string         `json:"method"`
	InputParams map[string]any `json:"inputs"`
}

type ServiceHandle func(evt *event.Event) *ServiceRequest

// ContractParserOpts - options for contract parser
// You can set hooks for pre-call and after-call
// For example, you can add logging or validation
type ContractParserOpts struct {
	PreCallHook   func(ctx mxbot.Ctx, req *ServiceRequest) (string, int, error)
	AfterCallHook func(ctx mxbot.Ctx, req *ServiceRequest, resp *contracts.MethodResponse) (string, int, error)
}

// SetContractParser - set contract parser. It is used for parsing events to service requests
// You can add hooks for pre-call and after-call with ContractParserOpts
func (bx *Bobrix) SetContractParser(
	parser func(evt *event.Event) *ServiceRequest,
	opts ...ContractParserOpts,
) {
	var opt ContractParserOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	bx.Use(
		mxbot.NewMessageHandler(
			func(ctx mxbot.Ctx) error {
				req := parser(ctx.Event())

				bx.logger.Info("incoming request",
					"event_id", ctx.Event().ID.String(),
					"room_id", ctx.Event().RoomID.String(),
					"service_id", req.ServiceID,
					"service_name", req.ServiceName,
					"method", req.MethodName,
				)

				ids := make([]string, 0, len(bx.servicesByID))
				for id, s := range bx.servicesByID {
					ids = append(ids, id.String()+"("+s.Service.Name+")")
				}
				bx.logger.Info("registered services", "count", len(ids), "services", ids)

				if raw := ctx.Event().Content.Raw; raw != nil {
					if p, ok := raw[BobrixPromptTag]; ok {
						bx.logger.Info("raw bobrix.prompt", "prompt", p)
					}
				}

				// ignore non-contract events
				if req == nil {
					return nil
				}

				// must have at least service_id or service name
				if req.ServiceID == "" && req.ServiceName == "" {
					return nil
				}

				if ctx.IsHandled() {
					return nil
				}
				if !ctx.TryClaim() {
					return nil
				}

				// --- Resolve service: prefer ServiceID, fallback to ServiceName ---
				var (
					svc *BobrixService
					ok  bool
				)

				if req.ServiceID != "" {
					id, err := uuid.Parse(req.ServiceID)
					if err != nil {
						return ctx.ErrorAnswer(
							fmt.Sprintf("Invalid service_id %q", req.ServiceID),
							contracts.ErrCodeBadRequest,
						)
					}

					svc, ok = bx.GetServiceByID(id)
					if !ok && req.ServiceName != "" {
						bx.logger.Warn("service_id not found, fallback to name",
							"service_id", req.ServiceID,
							"service_name", req.ServiceName,
						)
						svc, ok = bx.GetServiceByName(req.ServiceName)
					}
				} else {
					svc, ok = bx.GetServiceByName(req.ServiceName)
					if !ok {
						bx.logger.Error("service not found", "service", req.ServiceName)
						return ctx.ErrorAnswer(
							fmt.Sprintf("Service %q not found", req.ServiceName),
							contracts.ErrCodeServiceNotFound,
						)
					}
				}

				// service offline
				if !svc.IsOnline {
					svc.Handler(ctx, &contracts.MethodResponse{
						Err: fmt.Errorf("Service %q is offline", svc.Service.Name),
					}, nil)
					return nil
				}

				callOpts := contracts.CallOpts{}
				if thread := ctx.Thread(); thread != nil {
					data := ConvertThreadToMessages(thread, ctx.Bot().FullName())
					callOpts.Messages = data
				}

				if opt.PreCallHook != nil {
					errMsg, errCode, err := opt.PreCallHook(ctx, req)
					if err != nil {
						if err := ctx.ErrorAnswer(errMsg, errCode); err != nil {
							return err
						}
						return err
					}
				}

				resp, err := svc.Service.CallMethod(
					ctx.Context(),
					req.MethodName,
					req.InputParams,
					callOpts,
				)
				if err != nil {
					switch {
					case errors.Is(err, contracts.ErrMethodNotFound):
						if err := ctx.ErrorAnswer(
							fmt.Sprintf("Method %q not found", req.MethodName),
							contracts.ErrCodeMethodNotFound,
						); err != nil {
							return err
						}
					default:
						// resp may be nil if CallMethod returned error before response creation
						errCode := contracts.ErrCodeInternalServiceError
						if resp != nil && resp.ErrCode != 0 {
							errCode = resp.ErrCode
						}
						if err := ctx.ErrorAnswer(err.Error(), errCode); err != nil {
							return err
						}
					}
					return nil
				}

				if opt.AfterCallHook != nil {
					errMsg, errCode, err := opt.AfterCallHook(ctx, req, resp)
					if err != nil {
						if err := ctx.ErrorAnswer(errMsg, errCode); err != nil {
							return err
						}
						return err
					}
				}

				svc.Handler(ctx, resp, nil)
				return nil
			},
		),
	)
}

func ConvertThreadToMessages(thread *mxbot.MessagesThread, botName string) contracts.Messages {
	msgs := make([]map[contracts.ChatRole]string, len(thread.Messages))

	for i, msg := range thread.Messages {
		msgs[i] = map[contracts.ChatRole]string{}

		body := msg.Content.AsMessage().Body

		if msg.Sender.String() == botName {
			msgs[i][contracts.AssistantRole] = body
		} else {
			msgs[i][contracts.UserRole] = body
		}
	}

	return msgs
}
