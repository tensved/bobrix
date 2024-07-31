package contracts

// Input - describes the input data of the method
type Input struct {
	Name         string
	Description  string
	DefaultValue any
	IsRequired   bool
}

// Output - describes the output data of the method
type Output struct {
	Name         string
	Description  string
	DefaultValue any
}
