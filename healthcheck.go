package bobrix

import (
	"context"
	"sync"
	"time"
)

const (
	healthOk    = "healthy"   // used if service is healthy
	healthError = "unhealthy" // used if service is unhealthy (when ping returns an error)
)

// Health - status of bot and service
type Health struct {
	LastChecked time.Time `json:"last_checked"`
	Status      string    `json:"status" enums:"healthy,unhealthy"` // "healthy" or "unhealthy"
	Error       string    `json:"error,omitempty"`                  // only if status is "unhealthy". contains the error message
}

// BobrixStatus - status of the bot. Contains the health of the bot and the health of the services
// type BobrixStatus struct {
// 	MatrixStatus Health            `json:"matrix"`
// 	Services     map[string]Health `json:"services"`
// 	Health
// }

type ServiceStatus struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	Health      Health `json:"health"`
	Online      bool   `json:"online"`
}

type BobrixStatus struct {
	MatrixStatus Health                   `json:"matrix_status"`
	Services     map[string]ServiceStatus `json:"services"` // key = service_id
	Health       Health                   `json:"health"`
}

// Subscriber - subscriber for healthcheck.
// Receives BobrixStatus updates
// Can be used to subscribe to healthcheck updates
type Subscriber struct {
	dataChan chan *BobrixStatus
}

// notify - sends BobrixStatus to subscriber
func (s *Subscriber) notify(status *BobrixStatus) {
	s.dataChan <- status
}

// Read - reads single BobrixStatus from subscriber. Will block until data is available
func (s *Subscriber) Read() *BobrixStatus {
	return <-s.dataChan
}

// Close - closes subscriber. No more updates will be sent
func (s *Subscriber) Close() {
	close(s.dataChan)
}

// Sync - returns channel with BobrixStatus updates
// Can be used to subscribe to healthcheck updates
func (s *Subscriber) Sync() <-chan *BobrixStatus {
	return s.dataChan
}

// NewSubscriber - creates new subscriber. Used to subscribe to healthcheck updates
func NewSubscriber() *Subscriber {
	return &Subscriber{
		dataChan: make(chan *BobrixStatus),
	}
}

// Healthcheck - healthcheck interface for bobrix.
type Healthcheck interface {
	Subscribe() *Subscriber      // subscribe to healthcheck updates
	Unsubscribe(sub *Subscriber) // unsubscribe from healthcheck updates

	GetHealth() *BobrixStatus // get current healthcheck status
}

var _ Healthcheck = (*DefaultHealthcheck)(nil)

// DefaultHealthcheck - default healthcheck implementation
type DefaultHealthcheck struct {
	bobrix *Bobrix // reference to the bot

	subscribers []*Subscriber // subscribers that will receive healthcheck updates
	mx          *sync.RWMutex // mutex for subscribers

	interval time.Duration // healthcheck interval

	isAutoSwitch bool // switch offline/online status for services automatically based on health
}

// HealthcheckOption - options for healthcheck. Used in NewHealthcheck
type HealthcheckOption func(h *DefaultHealthcheck)

// WithInterval - set healthcheck interval. Default value described in defaultInterval
func WithInterval(interval time.Duration) HealthcheckOption {
	return func(h *DefaultHealthcheck) {
		h.interval = interval
	}
}

// WithAutoSwitch - turn on option to switch offline/online status for services automatically based on health
func WithAutoSwitch() HealthcheckOption {
	return func(h *DefaultHealthcheck) {
		h.isAutoSwitch = true
	}
}

const (
	defaultInterval = 60 * time.Second // default healthcheck interval
)

// NewHealthcheck - healthcheck constructor
// bobrix - reference to the bot
// opts - options for healthcheck
func NewHealthcheck(bobrix *Bobrix, opts ...HealthcheckOption) *DefaultHealthcheck {
	healthcheck := &DefaultHealthcheck{
		bobrix:       bobrix,
		subscribers:  make([]*Subscriber, 0),
		mx:           &sync.RWMutex{},
		interval:     defaultInterval,
		isAutoSwitch: false,
	}

	for _, opt := range opts {
		opt(healthcheck)
	}

	return healthcheck
}

// Subscribe - subscribe to healthcheck updates
// Returns subscriber that can be used to check healthcheck updates
func (h *DefaultHealthcheck) Subscribe() *Subscriber {
	h.mx.Lock()
	defer h.mx.Unlock()

	subscriber := NewSubscriber()
	h.subscribers = append(h.subscribers, subscriber)

	// start healthcheck if added first subscriber
	if len(h.subscribers) == 1 {
		go h.goHealthCheck()
	}

	return subscriber
}

// Unsubscribe - unsubscribe from healthcheck updates
// sub - subscriber to unsubscribe
func (h *DefaultHealthcheck) Unsubscribe(sub *Subscriber) {
	h.mx.Lock()
	defer h.mx.Unlock()

	for i, subscriber := range h.subscribers {
		if subscriber == sub {
			h.subscribers = append(h.subscribers[:i], h.subscribers[i+1:]...)
			break
		}
	}
}

// goHealthCheck - runs healthcheck in a loop
// Sends healthcheck updates to all subscribers
// Stops when there are no subscribers
func (h *DefaultHealthcheck) goHealthCheck() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for range ticker.C {

		// stop if there are no subscribers.
		// it may happen if healthcheck is unsubscribed while running
		if len(h.subscribers) == 0 {
			return
		}

		for _, sub := range h.subscribers {
			sub.notify(h.GetHealth())
		}
	}

}

// GetHealth - returns current healthcheck status
// Returns healthcheck status
func (h *DefaultHealthcheck) GetHealth() *BobrixStatus {
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	var botStatus Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		botStatus = h.getBotStatus(ctx)
	}()

	services := h.bobrix.Services()
	svcLength := len(services)

	serviceStatuses := make(map[string]ServiceStatus, svcLength)
	mx := &sync.Mutex{}

	wg.Add(svcLength)
	for _, service := range services {
		go func(service *BobrixService) {
			defer wg.Done()

			health := h.getServiceHealth(ctx, service)
			id := service.Service.ID.String()

			mx.Lock()
			serviceStatuses[id] = ServiceStatus{
				ServiceID:   id,
				ServiceName: service.Service.Name,
				Health:      health,
				Online:      service.IsOnline, // временно, обновим ниже если auto-switch
			}
			mx.Unlock()
		}(service)
	}

	wg.Wait()

	// --- Auto-switch: принимаем решение один раз, после сбора всех health ---
	if h.isAutoSwitch {
		anyServiceError := false
		for _, st := range serviceStatuses {
			if st.Health.Status == healthError {
				anyServiceError = true
				break
			}
		}

		// Обновляем IsOnline у сервисов и отражаем это в serviceStatuses
		for _, service := range services {
			id := service.Service.ID.String()

			st := serviceStatuses[id]
			if st.Health.Status == healthOk {
				service.IsOnline = true
			} else {
				service.IsOnline = false
			}
			st.Online = service.IsOnline
			serviceStatuses[id] = st
		}

		// Статус бота: online только если все сервисы ok
		if anyServiceError {
			h.bobrix.bot.SetIdleStatus()
		} else {
			h.bobrix.bot.SetOnlineStatus()
		}
	}

	// --- Итоговый health Bobrix ---
	bobrixHealth := Health{
		LastChecked: time.Now(),
		Status:      healthOk,
	}

	if botStatus.Status == healthError {
		bobrixHealth.Status = healthError
	}

	for _, svc := range serviceStatuses {
		if svc.Health.Status == healthError {
			bobrixHealth.Status = healthError
			break
		}
	}

	return &BobrixStatus{
		MatrixStatus: botStatus,
		Services:     serviceStatuses,
		Health:       bobrixHealth,
	}
}

// func (h *DefaultHealthcheck) GetHealth() *BobrixStatus {
// 	ctx := context.Background()
// 	wg := &sync.WaitGroup{}

// 	var botStatus Health
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		botStatus = h.getBotStatus(ctx)
// 	}()

// 	services := h.bobrix.Services()
// 	svcLength := len(services)

// 	// Лучше бы map[uuid.UUID]Health, но оставляю map[string]Health, чтобы меньше ломать API.
// 	// Ключ = service_id
// 	serviceStatuses := make(map[string]Health, svcLength)
// 	mx := &sync.Mutex{}

// 	wg.Add(svcLength)
// 	for _, service := range services {
// 		go func(service *BobrixService) {
// 			defer wg.Done()

// 			health := h.getServiceHealth(ctx, service)

// 			// auto-switch: обновляем флаг прямо в объекте сервиса
// 			if h.isAutoSwitch {
// 				if health.Status == healthOk {
// 					service.IsOnline = true
// 					h.bobrix.bot.SetOnlineStatus()
// 				} else {
// 					service.IsOnline = false
// 					h.bobrix.bot.SetIdleStatus()
// 				}
// 			}

// 			mx.Lock()
// 			// ключуем по ID, а не по имени
// 			serviceStatuses[service.Service.ID.String()] = health
// 			mx.Unlock()
// 		}(service)
// 	}

// 	wg.Wait()

// 	bobrixHealth := Health{
// 		LastChecked: time.Now(),
// 		Status:      healthOk,
// 	}

// 	if botStatus.Status == healthError {
// 		bobrixHealth.Status = healthError
// 	}

// 	for _, svc := range serviceStatuses {
// 		if svc.Status == healthError {
// 			bobrixHealth.Status = healthError
// 			break
// 		}
// 	}

// 	return &BobrixStatus{
// 		MatrixStatus: botStatus,
// 		Services:     serviceStatuses,
// 		Health:       bobrixHealth,
// 	}
// }

func (h *DefaultHealthcheck) getBotStatus(ctx context.Context) Health {
	status := Health{
		LastChecked: time.Now(),
		Status:      healthOk,
	}

	if err := h.bobrix.bot.Ping(ctx); err != nil {
		status = Health{
			LastChecked: time.Now(),
			Status:      healthError,
			Error:       err.Error(),
		}
	}

	return status
}

func (h *DefaultHealthcheck) getServiceHealth(ctx context.Context, service *BobrixService) Health {
	status := Health{
		LastChecked: time.Now(),
		Status:      healthOk,
	}

	if err := service.Service.Ping(ctx); err != nil {
		status = Health{
			LastChecked: time.Now(),
			Status:      healthError,
			Error:       err.Error(),
		}
	}

	return status
}
