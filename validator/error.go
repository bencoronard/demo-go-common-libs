package validator

type ValidationError interface {
	Data() []FieldValidationError
	Error() string
}

type validationError struct {
	errors []FieldValidationError
}

func (ve *validationError) Error() string {
	return "invalid input"
}

func (ve *validationError) Data() []FieldValidationError {
	return ve.errors
}
