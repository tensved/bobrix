package bobrix

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Engine is the main component of the library
// Bots can be attached to it (see Bobrix).
// It is also responsible for launching all bots
type Engine struct {
	bots     []*Bobrix
	services []*BobrixService
	mx       *sync.RWMutex

	logger *slog.Logger
}

// NewEngine - Bobrix engine constructor
func NewEngine() *Engine {
	return &Engine{
		bots:     make([]*Bobrix, 0),
		services: make([]*BobrixService, 0),
		mx:       &sync.RWMutex{},
		logger:   slog.Default(),
	}
}

// ConnectBot - add bot to the engine. It is used for adding bots to the engine
func (e *Engine) ConnectBot(bot *Bobrix) {
	e.mx.Lock()
	e.bots = append(e.bots, bot)
	e.mx.Unlock()
}

// ConnectService - add service to the store in engine.
func (e *Engine) ConnectService(service *BobrixService) {
	e.mx.Lock()
	e.services = append(e.services, service)
	e.mx.Unlock()
}

// Run - launch all bots.
// It uses semaphore to limit the number of bots that can start (login) at the same time
func (e *Engine) Run(ctx context.Context) error {

	semaphore := make(chan struct{}, 5)

	for _, bot := range e.bots {

		go func(bot *Bobrix) {
			semaphore <- struct{}{}
			go func() {
				time.Sleep(2 * time.Second)
				<-semaphore
			}()
			ctx := context.Background()
			if err := bot.Run(ctx); err != nil {
				e.logger.Error("failed to run bot", "error", err)
			}
		}(bot)
	}

	<-ctx.Done()

	return ctx.Err()
}

// Stop - stop all bots
func (e *Engine) Stop(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(e.bots))
	for _, bot := range e.bots {
		go func(bot *Bobrix) {
			ctx := context.Background()
			if err := bot.Stop(ctx); err != nil {
				e.logger.Error("failed to stop bot", "error", err)
			}
			wg.Done()
		}(bot)
	}

	wg.Wait()

	return nil
}

// Bots - return all bots
func (e *Engine) Bots() []*Bobrix {
	e.mx.RLock()
	defer e.mx.RUnlock()

	return e.bots
}

// GetBot - return bot by name. If the bot is not found, it returns nil
func (e *Engine) GetBot(name string) *Bobrix {
	e.mx.RLock()
	defer e.mx.RUnlock()

	for _, bot := range e.bots {
		if bot.Name() == name {
			return bot
		}
	}
	return nil
}

// Services - return all services
func (e *Engine) Services() []*BobrixService {
	e.mx.RLock()
	defer e.mx.RUnlock()

	return e.services
}

// GetService - return service by name. If the service is not found, it returns nil
func (e *Engine) GetService(name string) *BobrixService {
	e.mx.RLock()
	defer e.mx.RUnlock()

	for _, service := range e.services {
		if service.Service.Name == name {
			return service
		}
	}

	return nil
}
