package contracts

type IOType string

const (
	IOTypeText  IOType = "text"
	IOTypeAudio IOType = "audio"
	IOTypeImage IOType = "image"
	IOTypeVideo IOType = "video"
	IOTypeFile  IOType = "file"
)

// Input - describes the input data of the method
type Input struct {
	Name         string `json:"name" yaml:"name"`
	Type         IOType `json:"type" yaml:"type"` // text, audio, image, video, file
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	DefaultValue any    `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	IsRequired   bool   `json:"is_required" yaml:"is_required"`
	value        any
}

func (i *Input) SetValue(value any) {
	i.value = value
}

func (i *Input) Value() any {
	return i.value
}

// Output - describes the output data of the method
type Output struct {
	Name         string `json:"name" yaml:"name"`
	Type         IOType `json:"type" yaml:"type"` // text, audio, image, video, file
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	DefaultValue any    `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	value        any
}

func (o *Output) SetValue(value any) {
	o.value = value
}

func (o *Output) Value() any {
	return o.value
}
