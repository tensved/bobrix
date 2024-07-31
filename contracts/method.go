package contracts

import (
	"errors"
	"fmt"
)

// MethodResponse - describes the response of the method
type MethodResponse struct {
	Data  map[string]any
	Error error
}

// MethodHandler - describes the handler of the method
type MethodHandler interface {
	Do(inputData map[string]any) *MethodResponse
}

// Method - describes the method of the service
// Contains the name, description, inputs and outputs.
// Also contains the name of the handler function and the handler itself
type Method struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Inputs  []*Input  `json:"inputs"`
	Outputs []*Output `json:"outputs"`

	HandlerName string        `json:"handler_name"`
	Handler     MethodHandler `json:"-"`
}

// Call - calls the method
// If the method has no handler, it returns an error
// Otherwise, it calls the handler
// It returns the result of the handler
func (m *Method) Call(inputData map[string]any) *MethodResponse {
	if m.Handler == nil {
		return &MethodResponse{
			Error: fmt.Errorf("%w: %s", ErrHandlerNotFound, m.HandlerName),
		}
	}
	return m.Handler.Do(inputData)
}

// checkInputs - checks the inputs of the method
// If the input is required and not present, it returns an error
// If the input is not required and not present, it sets the default value
func (m *Method) checkInputs(inputData map[string]any) error {

	var err error

	for _, methodInput := range m.Inputs {
		_, ok := inputData[methodInput.Name]
		if !ok {
			if methodInput.DefaultValue != nil {
				inputData[methodInput.Name] = methodInput.DefaultValue
				continue
			}

			if methodInput.IsRequired {
				err = errors.Join(err, fmt.Errorf("%w: %s", ErrInputRequired, methodInput.Name))
			}
		}

		// TODO: another checks
	}

	return err
}

func (m *Method) ValidateInputs(inputData map[string]any) error {
	return m.checkInputs(inputData)
}
