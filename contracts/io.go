package contracts

// Input - describes the input data of the method
type Input struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // text, audio, image, video
	Description  string `json:"description,omitempty"`
	DefaultValue any    `json:"default_value,omitempty"`
	IsRequired   bool   `json:"is_required"`
}

// Output - describes the output data of the method
type Output struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // text, audio, image, video
	Description  string `json:"description,omitempty"`
	DefaultValue any    `json:"default_value,omitempty"`
}
