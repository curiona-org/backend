package validator

// Validator is an interface that defines methods for validating structs and parsing validation errors.
// Implementations of this interface should provide the logic for validating a struct and converting
// validation errors into a slice of ErrorMessage.
type Validator interface {
	// Validate takes any struct as input and returns an error if validation fails.
	Validate(s any) error
	// ParseErrors takes an error returned by Validate and converts it into a slice of ErrorMessage.
	ParseErrors(err error) []ErrorMessage
}

var _ Validator = (*PlaygroundValidator)(nil)

// ErrorMessage contains the field name and the error message for a validation error.
// Useful for returning validation errors in a response.
type ErrorMessage struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
