package contracts

import (
	"context"
	"fmt"
)

// Service - describes the service of the application
// Contains the name, description and methods
// Can be used to call methods
type Service struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Methods map[string]*Method `json:"methods" yaml:"methods"`

	Pinger *Ping `json:"pinger,omitempty" yaml:"pinger,omitempty"`
}

// CallMethod - calls the method with the given name
// If the method does not exist, it returns an error
// Otherwise, it calls the method and returns the result
func (s *Service) CallMethod(methodName string, inputData map[string]any) (*MethodResponse, error) {
	method, ok := s.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMethodNotFound, methodName)
	}

	return method.Call(inputData), nil
}

func (s *Service) AddMethod(method *Method) {
	s.Methods[method.Name] = method
}

// Ping - pings the service
// If the service does not have a pinger, it returns an error
// Otherwise, it calls the pinger and returns the result
// If the pinger returns nil, it means that the service is ok
func (s *Service) Ping(ctx context.Context) error {
	if s.Pinger == nil {
		return ErrNoHealthCheckProvided
	}

	return s.Pinger.Do(ctx)
}
