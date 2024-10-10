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

type CallOpts struct {
	Messages Messages
}

// CallMethod - calls the method with the given name
// If the method does not exist, it returns an error
// Otherwise, it calls the method and returns the result
func (s *Service) CallMethod(methodName string, inputData map[string]any, opts ...CallOpts) (*MethodResponse, error) {
	method, ok := s.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMethodNotFound, methodName)
	}

	return method.Call(inputData, opts...)
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

// GetDefaultMethod - returns the method with the default flag
// If the service does not have a default method, it returns nil
func (s *Service) GetDefaultMethod() *Method {
	for _, method := range s.Methods {
		if method.IsDefault {
			return method
		}
	}

	return nil
}

// ServicePublic - describes the service as public.
// It can be used to print information about the service without sensitive data
type ServicePublic struct {
	Name        string                  `json:"name" yaml:"name"`
	Description string                  `json:"description,omitempty" yaml:"description,omitempty"`
	Methods     map[string]MethodPublic `json:"methods" yaml:"methods"`
}

// AsPublic - returns the service as a public service
func (s *Service) AsPublic() ServicePublic {

	var methodsPublic map[string]MethodPublic

	if s.Methods != nil {
		methodsPublic = make(map[string]MethodPublic, len(s.Methods))
		for _, method := range s.Methods {
			methodsPublic[method.Name] = method.AsPublic()
		}
	}

	return ServicePublic{
		Name:        s.Name,
		Description: s.Description,
		Methods:     methodsPublic,
	}
}
