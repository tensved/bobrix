package contracts

// MethodResponse - describes the response of the method
type MethodResponse struct {
	Data map[string]any
	Err  error
}

type HandlerFunc func(inputData map[string]any) *MethodResponse

// Handler - describes the handler of the method
type Handler struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Args        map[string]any `json:"args"`

	Do HandlerFunc `json:"-"`
}
