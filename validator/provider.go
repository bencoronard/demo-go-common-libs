package validator

import (
	val "github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i any) error
}

type validator struct {
	validator *val.Validate
}

func (v *validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func New() (Validator, error) {
	v := val.New(val.WithRequiredStructEnabled())

	v.RegisterValidation("notblank", notblank)

	return &validator{
		validator: v,
	}, nil
}
