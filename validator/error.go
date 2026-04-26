package validator

type ValidationError struct {
	Errors []FieldValidationError
}

func (ve *ValidationError) Error() string {
	return "invalid input"
}
