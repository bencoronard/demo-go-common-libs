package validator

import (
	"fmt"

	val "github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i any) error
}

func New() (Validator, error) {
	v := val.New(val.WithRequiredStructEnabled())

	if err := v.RegisterValidation("notblank", notblank); err != nil {
		return nil, fmt.Errorf("failed to register validation: %w", err)
	}

	return &validator{validator: v}, nil
}
