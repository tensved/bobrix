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
	Status string `json:"status"`          // "healthy" or "unhealthy"
	Error  string `json:"error,omitempty"` // only if status is "unhealthy". contains the error message
}

// ServiceStatus - status of service
// running - true if service is online.
type ServiceStatus struct {
	Health  Health `json:"health"`
	Running bool   `json:"running"`
}

// BobrixStatus - status of the bot. Contains the health of the bot and the health of the services
type BobrixStatus struct {
	Running   bool                     `json:"running"`
	BotStatus Health                   `json:"bot"`
	Services  map[string]ServiceStatus `json:"services"`
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
	defaultInterval = 5 * time.Second // default healthcheck interval
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
		botStatus = h.getBotStatus(ctx)
		wg.Done()
	}()

	svcLength := len(h.bobrix.services) // length of services array

	serviceStatuses := make(map[string]ServiceStatus, svcLength)
	serviceMX := &sync.RWMutex{}

	wg.Add(svcLength) // add wg for each service

	for i, service := range h.bobrix.services {

		go func(i int, service *BobrixService) {

			health := h.getServiceHealth(ctx, service)

			// if isAutoSwitch is enabled, update service status based on health
			if h.isAutoSwitch {
				h.bobrix.services[i].IsOnline = health.Status == healthOk
			}

			serviceMX.Lock()
			serviceStatuses[service.Service.Name] = ServiceStatus{
				Health:  health,
				Running: service.IsOnline,
			}
			serviceMX.Unlock()

			wg.Done()
		}(i, service)
	}

	wg.Wait() // wait for all pings to finish (bot and services)

	return &BobrixStatus{
		Running:   true,
		BotStatus: botStatus,
		Services:  serviceStatuses,
	}
}

func (h *DefaultHealthcheck) getBotStatus(ctx context.Context) Health {
	status := Health{
		Status: healthOk,
	}

	if err := h.bobrix.bot.Ping(ctx); err != nil {
		status = Health{
			Status: healthError,
			Error:  err.Error(),
		}
	}

	return status
}

func (h *DefaultHealthcheck) getServiceHealth(ctx context.Context, service *BobrixService) Health {
	status := Health{
		Status: healthOk,
	}

	if err := service.Service.Ping(ctx); err != nil {
		status = Health{
			Status: healthError,
			Error:  err.Error(),
		}
	}

	return status
}
