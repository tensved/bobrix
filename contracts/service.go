package contracts

import "fmt"

// Service - describes the service of the application
// Contains the name, description and methods
// Can be used to call methods
type Service struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Methods map[string]*Method `json:"methods"`
}

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
