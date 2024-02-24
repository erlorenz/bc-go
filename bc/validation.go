package bc

import (
	"errors"
)

var (
	ErrorEmptyString = errors.New("empty string")
)

type Validator interface {
	Validate() error
}

func stringNotEmpty(s string) error {
	if s == "" {
		return ErrorEmptyString
	}
	return nil
}
