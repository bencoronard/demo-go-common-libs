package utility

import (
	val "github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i any) error
}

type validator struct {
	validator *val.Validate
}

func NewValidator() (Validator, error) {
	return &validator{
		validator: val.New(val.WithRequiredStructEnabled()),
	}, nil
}

func (v *validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
