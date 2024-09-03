package contracts

// IOType represents the type of input or output in a method (e.g., text, audio, image).
type IOType string

// Constants for different IOType values.
const (
	IOTypeText    IOType = "text"
	IOTypeNumber  IOType = "number"
	IOTypeBoolean IOType = "boolean"
	IOTypeAudio   IOType = "audio"
	IOTypeImage   IOType = "image"
	IOTypeVideo   IOType = "video"
	IOTypeFile    IOType = "file"
	IOTypeJSON    IOType = "json"
)

// Input represents the input data of a method.
type Input struct {
	Name         string `json:"name" yaml:"name"`                                   // Name of the input.
	Type         IOType `json:"type" yaml:"type"`                                   // Type of the input (e.g., text, audio, image, etc.).
	Description  string `json:"description,omitempty" yaml:"description,omitempty"` // Optional description of the input.
	DefaultValue any    `json:"default,omitempty" yaml:"default,omitempty"`         // Optional default value for the input.
	IsRequired   bool   `json:"is_required" yaml:"is_required"`                     // Indicates if the input is required.
	value        any    // Internal value of the input.
}

// SetValue sets the internal value of the input.
func (i *Input) SetValue(value any) {
	i.value = value
}

// Value returns the internal value of the input.
func (i *Input) Value() any {
	return i.value
}

// Output represents the output data of a method.
type Output struct {
	Name         string `json:"name" yaml:"name"`                                   // Name of the output.
	Type         IOType `json:"type" yaml:"type"`                                   // Type of the output (e.g., text, audio, image, etc.).
	Description  string `json:"description,omitempty" yaml:"description,omitempty"` // Optional description of the output.
	DefaultValue any    `json:"default,omitempty" yaml:"default,omitempty"`         // Optional default value for the output.
	value        any
}

// SetValue sets the internal value of the output.
func (o *Output) SetValue(value any) {
	o.value = value
}

// Value returns the internal value of the output.
func (o *Output) Value() any {
	return o.value
}
