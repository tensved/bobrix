package contracts

import (
	"context"
	"errors"
	"fmt"
)

// Method - describes the method of the service
// Contains the name, description, inputs and outputs.
// Also contains the name of the handler function and the handler itself
type Method struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Inputs  []Input  `json:"inputs" yaml:"inputs"`
	Outputs []Output `json:"outputs" yaml:"outputs"`

	Handler *Handler `json:"handler" yaml:"handler"`
}

// Call - calls the method
// If the method has no handler, it returns an error
// Otherwise, it calls the handler
// It returns the result of the handler
// Preferably use CallWithContext
func (m *Method) Call(inputData map[string]any) (*MethodResponse, error) {
	ctx := context.Background()
	return m.CallWithContext(ctx, inputData)
}

// CallWithContext - calls the method with the given context
// If the method has no handler, it returns an error
// Otherwise, it calls the handler
// It returns the result of the handler
func (m *Method) CallWithContext(ctx context.Context, inputData map[string]any) (*MethodResponse, error) {
	if m.Handler == nil {
		return nil, fmt.Errorf("%w: %s", ErrHandlerNotFound, m.Name)
	}

	inputs, err := m.processInputs(inputData)
	if err != nil {
		return nil, err
	}

	// prepare outputs for the handler context
	outputs := make(map[string]Output, len(m.Outputs))
	for _, output := range m.Outputs {
		outputs[output.Name] = output
	}

	c := NewHandlerContext(ctx, inputs, outputs) // create new handler context with the processed inputs

	err = m.Handler.Do(c)

	response := &MethodResponse{
		Outputs: c.Outputs(),
		Err:     err,
	}

	return response, nil

}

// GetInputs - returns the list of inputs
func (m *Method) GetInputs() []Input {
	return m.Inputs
}

// InputFilter - filter for inputs. Returns true if the input should be included
type InputFilter func(input Input) bool

// GetInputsWithFilter - returns the list of inputs with the given filter.
func (m *Method) GetInputsWithFilter(filter InputFilter) []Input {
	var inputs []Input
	for _, input := range m.Inputs {
		if filter(input) {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// GetOutputs - returns the list of outputs
func (m *Method) GetOutputs() []Output {
	return m.Outputs
}

// GetInput - returns the input with the given name
func (m *Method) GetInput(name string) (Input, bool) {
	for _, input := range m.Inputs {
		if input.Name == name {
			return input, true
		}
	}
	return Input{}, false
}

// GetOutput - returns the output with the given name
func (m *Method) GetOutput(name string) (Output, bool) {
	for _, output := range m.Outputs {
		if output.Name == name {
			return output, true
		}
	}
	return Output{}, false
}

// GetTextInputs - returns the list of text inputsq
func (m *Method) GetTextInputs() []Input {
	var inputs []Input
	for _, input := range m.Inputs {
		if input.Type == IOTypeText {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// GetAudioInputs - returns the list of audio inputs
func (m *Method) GetAudioInputs() []Input {
	var inputs []Input
	for _, input := range m.Inputs {
		if input.Type == IOTypeAudio {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// GetImageInputs - returns the list of image inputs
func (m *Method) GetImageInputs() []Input {
	var inputs []Input
	for _, input := range m.Inputs {
		if input.Type == IOTypeImage {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// GetVideoInputs - returns the list of video inputs
func (m *Method) GetVideoInputs() []Input {
	var inputs []Input
	for _, input := range m.Inputs {
		if input.Type == IOTypeVideo {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// GetFileInputs - returns the list of file inputs
func (m *Method) GetFileInputs() []Input {
	var inputs []Input
	for _, input := range m.Inputs {
		if input.Type == IOTypeFile {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// processInputs - checks the inputs of the method and fills values with default values
// If the input is required and not present, it returns an error
// If the input is not required and not present, it sets the default value
func (m *Method) processInputs(inputData map[string]any) (map[string]Input, error) {

	result := make(map[string]Input, len(m.Inputs))

	var err error

	for _, methodInput := range m.Inputs {

		userInput, ok := inputData[methodInput.Name]
		if !ok { // if the input is not present, check if it is required

			// if input is not present and has no default value, check if it is required. If it is, return an error
			if methodInput.IsRequired && methodInput.DefaultValue == nil {
				err = errors.Join(err, fmt.Errorf("%w: %s", ErrInputRequired, methodInput.Name))
				continue
			}

			// if input is not present and has a default value, set it
			userInput = methodInput.DefaultValue
		}

		methodInput.SetValue(userInput)        // set the value of the input
		result[methodInput.Name] = methodInput // add the input to the result
	}

	return result, err
}
