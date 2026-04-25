package validator

import (
	"errors"
	"reflect"
	"strings"

	val "github.com/go-playground/validator/v10"
)

type validator struct {
	validator *val.Validate
}

func (v *validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		var ve val.ValidationErrors

		if !errors.As(err, &ve) {
			return err
		}

		var data []FieldValidationError
		for _, fe := range ve {
			data = append(data, FieldValidationError{
				Field:   strings.ToLower(fe.Field()),
				Message: fe.Tag(),
			})
		}

		return &validationError{
			errors: data,
		}
	}
	return nil
}

func notblank(fl val.FieldLevel) bool {
	if fl.Field().Kind() == reflect.Pointer {
		if fl.Field().IsNil() {
			return false
		}
		return len(strings.TrimSpace(fl.Field().Elem().String())) > 0
	}
	return len(strings.TrimSpace(fl.Field().String())) > 0
}
