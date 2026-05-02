package validator

import (
	val "github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i any) error
}

func New() (Validator, error) {
	v := val.New(val.WithRequiredStructEnabled())

	if err := v.RegisterValidation("notblank", notblank); err != nil {
		return nil, err
	}

	return &validator{validator: v}, nil
}
