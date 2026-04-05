package validator

import (
	val "github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i any) error
}

func New() (Validator, error) {
	v := val.New(val.WithRequiredStructEnabled())

	v.RegisterValidation("notblank", notblank)

	return &validator{
		validator: v,
	}, nil
}
