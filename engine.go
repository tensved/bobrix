package bobrix

import (
	"context"
	"log/slog"
	"sync"
)

// Engine is the main component of the library
// Bots can be attached to it (see Bobrix).
// It is also responsible for launching all bots
type Engine struct {
	bots     []*Bobrix
	services []*BobrixService
	mx       *sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		bots:     make([]*Bobrix, 0),
		services: make([]*BobrixService, 0),
		mx:       &sync.RWMutex{},
	}
}

func (e *Engine) ConnectBot(bot *Bobrix) {
	e.mx.Lock()
	e.bots = append(e.bots, bot)
	e.mx.Unlock()
}

func (e *Engine) ConnectService(service *BobrixService) {
	e.mx.Lock()
	e.services = append(e.services, service)
	e.mx.Unlock()
}

func (e *Engine) Run(ctx context.Context) error {

	for _, bot := range e.bots {
		go func(bot *Bobrix) {
			ctx := context.Background()
			if err := bot.Run(ctx); err != nil {
				slog.Error("failed to run bot", "error", err)
			}
		}(bot)
	}

	<-ctx.Done()

	return ctx.Err()
}

func (e *Engine) Stop(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(e.bots))
	for _, bot := range e.bots {
		go func(bot *Bobrix) {
			ctx := context.Background()
			if err := bot.Stop(ctx); err != nil {
				slog.Error("failed to stop bot", "error", err)
			}
			wg.Done()
		}(bot)
	}

	wg.Wait()

	return nil
}

func (e *Engine) Bots() []*Bobrix {
	e.mx.RLock()
	defer e.mx.RUnlock()

	return e.bots
}

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

func (e *Engine) Services() []*BobrixService {
	e.mx.RLock()
	defer e.mx.RUnlock()

	return e.services
}

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
