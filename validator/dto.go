package validator

type FieldValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
