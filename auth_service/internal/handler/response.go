package handler

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

func ValidationErrors(errs validator.ValidationErrors) error {
	for _, err := range errs {
		if err.Tag() == "required_if" {
			return fmt.Errorf("you must fill the '%s' value", err.Field())
		}
	}
	return nil
}
