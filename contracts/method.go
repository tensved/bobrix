package contracts

import (
	"errors"
	"fmt"
)

// Method - describes the method of the service
// Contains the name, description, inputs and outputs.
// Also contains the name of the handler function and the handler itself
type Method struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Inputs  []*Input  `json:"inputs"`
	Outputs []*Output `json:"outputs"`

	Handler *Handler `json:"handler"`
}

// Call - calls the method
// If the method has no handler, it returns an error
// Otherwise, it calls the handler
// It returns the result of the handler
func (m *Method) Call(inputData map[string]any) *MethodResponse {
	if m.Handler == nil {
		return &MethodResponse{
			Error: fmt.Errorf("%w: %s", ErrHandlerNotFound, m.Handler.Name),
		}
	}
	return m.Handler.Do(inputData)
}

// GetInputs - returns the list of inputs
func (m *Method) GetInputs() []*Input {
	return m.Inputs
}

// GetOutputs - returns the list of outputs
func (m *Method) GetOutputs() []*Output {
	return m.Outputs
}

// GetInput - returns the input with the given name
func (m *Method) GetInput(name string) *Input {
	for _, input := range m.Inputs {
		if input.Name == name {
			return input
		}
	}
	return nil
}

// GetOutput - returns the output with the given name
func (m *Method) GetOutput(name string) *Output {
	for _, output := range m.Outputs {
		if output.Name == name {
			return output
		}
	}
	return nil
}

// GetTextInputs - returns the list of text inputsq
func (m *Method) GetTextInputs() []*Input {
	var inputs []*Input
	for _, input := range m.Inputs {
		if input.Type == "text" {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// GetAudioInputs - returns the list of audio inputs
func (m *Method) GetAudioInputs() []*Input {
	var inputs []*Input
	for _, input := range m.Inputs {
		if input.Type == "audio" {
			inputs = append(inputs, input)
		}
	}
	return inputs
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
