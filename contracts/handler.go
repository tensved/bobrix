package contracts

import "fmt"

// MethodResponse describes the response of a method, including output data and any errors.
type MethodResponse struct {
	Outputs     map[string]Output // Outputs contains the output data of the method.
	ServiceName string            // ServiceName is the name of the service that the method belongs to.
	Err         error             // Err contains any error encountered during method execution.
	ErrCode     int               // ErrCode is the error code of the method.
}

// Get retrieves the value of a specific output by name.
// It returns the value and a boolean indicating whether the output was found.
func (m *MethodResponse) Get(name string) (any, bool) {
	output, ok := m.Outputs[name]
	return output.Value(), ok
}

// GetString retrieves the string representation of a specific output by name.
// It returns the formatted value and a boolean indicating whether the output was found.
func (m *MethodResponse) GetString(name string) (string, bool) {

	output, ok := m.Outputs[name]

	if !ok {
		return "", false
	}

	return fmt.Sprintf("%v", output.Value()), true
}

// HandlerFunc defines the function signature for method handlers,
// which receive a HandlerContext and return an error if something goes wrong.
type HandlerFunc func(ctx HandlerContext) error

// Handler represents a method handler with metadata, arguments, and the handler function itself.
type Handler struct {
	Name        string         `json:"name"`                  // Name is the name of the handler.
	Description string         `json:"description,omitempty"` // Description is an optional description of the handler.
	Args        map[string]any `json:"args"`                  // Args are the arguments for the handler.

	Do HandlerFunc `json:"-"` // Do is the handler function itself, which is not serialized to JSON.
}
