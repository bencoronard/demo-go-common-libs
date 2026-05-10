package validator

import (
	"errors"
	"fmt"
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

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	ve, ok := errors.AsType[val.ValidationErrors](err)
	if !ok {
		return fmt.Errorf("failed to validate input: %w", err)
	}

	s := make([]FieldValidationError, 0, len(ve))
	for _, fe := range ve {
		field, _ := t.FieldByName(fe.StructField())

		msg := field.Tag.Get(fmt.Sprintf("%s:msg", fe.Tag()))
		if msg == "" {
			msg = fmt.Sprintf("%v is not valid", fe.Value())
		}

		s = append(s, FieldValidationError{
			Field:   strings.ToLower(fe.Field()),
			Message: msg,
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
