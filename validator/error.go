package validator

type ValidationError interface {
	Data() []FieldValidationError
	Error() string
}

type validationError struct {
	errors []FieldValidationError
}

func (ve *validationError) Error() string {
	return "data did not pass validations"
}

func (ve *validationError) Data() []FieldValidationError {
	return ve.errors
}
