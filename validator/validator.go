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
	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}

	ve, ok := errors.AsType[val.ValidationErrors](err)
	if !ok {
		return err
	}

	s := make([]FieldValidationError, 0, len(ve))
	for _, fe := range ve {
		s = append(s, FieldValidationError{
			Field:   strings.ToLower(fe.Field()),
			Message: fe.Tag(),
		})
	}

	return &ValidationError{Errors: s}
}

func notblank(fl val.FieldLevel) bool {
	if fl.Field().Kind() != reflect.Pointer {
		return len(strings.TrimSpace(fl.Field().String())) > 0
	}

	if fl.Field().IsNil() {
		return false
	}

	return len(strings.TrimSpace(fl.Field().Elem().String())) > 0
}
