package contracts

import (
	"context"
	"encoding/json"
	"fmt"
)

// HandlerContext defines an interface for managing inputs and outputs
// and providing access to the underlying context.
type HandlerContext interface {

	// Context returns the underlying context.
	Context() context.Context

	// Get retrieves the value of the specified input.
	Get(inputName string) any

	// GetString retrieves the string representation of the specified input.
	GetString(inputName string) (string, bool)

	// GetInt retrieves the integer representation of the specified input.
	GetInt(inputName string) (int, bool)

	// GetFloat retrieves the float representation of the specified input.
	GetFloat(inputName string) (float64, bool)

	// GetBool retrieves the boolean representation of the specified input.
	GetBool(inputName string) (bool, bool)

	// GetInput retrieves an Input by its name, if present.
	GetInput(name string) (Input, bool)

	// Inputs returns all input mappings.
	Inputs() map[string]Input

	// Set assigns a value to the specified output.
	Set(outputName string, value any)

	// GetOutput retrieves an Output by its name, if present.
	GetOutput(name string) (Output, bool)

	// Outputs returns all output mappings.
	Outputs() map[string]Output

	// JSON populates the outputs by marshaling and unmarshaling the provided data.
	JSON(data any) error

	Messages() Messages
	SetMessages(messages Messages)
}

// Ensure DefaultHandlerContext implements HandlerContext.
var _ HandlerContext = (*DefaultHandlerContext)(nil)

// DefaultHandlerContext is the default implementation of the HandlerContext interface.
type DefaultHandlerContext struct {
	ctx     context.Context
	inputs  map[string]Input
	outputs map[string]Output

	messages Messages
}

type HandlerContextOpts func(handlerContext HandlerContext)

func WithMessages(messages Messages) HandlerContextOpts {
	return func(handlerContext HandlerContext) {
		handlerContext.SetMessages(messages)
	}
}

// NewHandlerContext creates a new DefaultHandlerContext with the provided context, inputs, and outputs.
func NewHandlerContext(ctx context.Context, inputs map[string]Input, outputs map[string]Output, opts ...HandlerContextOpts) HandlerContext {
	handlerContext := &DefaultHandlerContext{
		ctx:      ctx,
		inputs:   inputs,
		outputs:  outputs,
		messages: Messages{},
	}

	for _, opt := range opts {
		opt(handlerContext)
	}

	return handlerContext
}

func (h *DefaultHandlerContext) Context() context.Context {
	return h.ctx
}

func (h *DefaultHandlerContext) Get(inputName string) any {
	inp, ok := h.GetInput(inputName)
	if !ok {
		return nil
	}

	return inp.Value()
}

func (h *DefaultHandlerContext) GetString(inputName string) (string, bool) {
	inp, ok := h.GetInput(inputName)
	if !ok || inp.Value() == nil {
		return "", false
	}

	return fmt.Sprintf("%v", inp.Value()), true
}

func (h *DefaultHandlerContext) GetInt(inputName string) (int, bool) {
	inp, ok := h.GetInput(inputName)
	if !ok {
		return 0, false
	}

	inpInt, ok := inp.Value().(float64)
	if !ok {
		return 0, false
	}

	return int(inpInt), true
}

func (h *DefaultHandlerContext) GetFloat(inputName string) (float64, bool) {
	inp, ok := h.GetInput(inputName)
	if !ok {
		return 0, false
	}

	inpFloat, ok := inp.Value().(float64)
	if !ok {
		return 0, false
	}

	return inpFloat, true
}

func (h *DefaultHandlerContext) GetBool(inputName string) (bool, bool) {
	inp, ok := h.GetInput(inputName)
	if !ok {
		return false, false
	}

	inpBool, ok := inp.Value().(bool)
	if !ok {
		return false, false
	}

	return inpBool, true
}

func (h *DefaultHandlerContext) GetInput(name string) (Input, bool) {
	inp, ok := h.inputs[name]
	return inp, ok
}

func (h *DefaultHandlerContext) Inputs() map[string]Input {
	return h.inputs
}

func (h *DefaultHandlerContext) Set(outputName string, value any) {
	out, ok := h.outputs[outputName]
	if ok {
		out.SetValue(value)
		h.outputs[outputName] = out
	}
}

func (h *DefaultHandlerContext) GetOutput(name string) (Output, bool) {
	out, ok := h.outputs[name]
	return out, ok
}

func (h *DefaultHandlerContext) Outputs() map[string]Output {
	return h.outputs
}

// JSON populates the outputs by marshaling the provided data to JSON and unmarshaling it into a map.
// If the input data is already a map, it is directly used to set outputs
func (h *DefaultHandlerContext) JSON(i any) error {
	if i == nil {
		return nil
	}

	outputsMap, ok := i.(map[string]any)

	if !ok {
		// Marshal the input data to JSON and unmarshal it into a map.
		bytes, err := json.Marshal(i)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(bytes, &outputsMap); err != nil {
			return err
		}
	}

	// Set each key-value pair in the outputs map.
	for k, v := range outputsMap {
		h.Set(k, v)
	}

	return nil
}

func (h *DefaultHandlerContext) Messages() Messages {

	return h.messages
}

func (h *DefaultHandlerContext) SetMessages(messages Messages) {
	h.messages = messages
}
