package v2_test

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	ErrorEmptyString = errors.New("empty string")
)

// Validator is an interface that wraps the Validate method.
// Validate returns an error and can use any validation logic.
type Validator interface {
	Validate() error
}

// Use this go-playground/validator validate session
var v = validator.New(validator.WithRequiredStructEnabled())

// ValidateStruct uses the validate instance to validate a struct
// and return an error. It calls the validate.Struct method and does
// a check for the InvalidValidationError.
func validateStruct(s any) error {
	err := v.Struct(s)

	// Return early if no error
	if err == nil {
		return nil
	}

	// this check is only needed when your code could produce an
	// invalid value for validation such as interface with nil value
	if _, exists := err.(*validator.InvalidValidationError); exists {
		return fmt.Errorf("invalid value passed to validation: %w", err)
	}

	var joinedErrs error

	for _, fieldError := range err.(validator.ValidationErrors) {
		err := fmt.Errorf("invalid %s: %s", fieldError.StructField(), fieldError.Error())
		joinedErrs = errors.Join(joinedErrs, err)
	}
	return joinedErrs
}

func stringNotEmpty(s string) error {
	if s == "" {
		return ErrorEmptyString
	}
	return nil
}
