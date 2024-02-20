package bc

import (
	"errors"
	"fmt"
	"strings"
)

// Validator returns an error from the Validate method
type Validator interface {
	Validate() error
}

// ValidatorMap is a map of keys and domaintypes that meet the
// Validator interface
type ValidatorMap map[string]Validator

func (m ValidatorMap) Validate() error {
	problems := []string{}
	for k, v := range m {
		if err := v.Validate(); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %s", k, err.Error()))
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, ", "))
	}
	return nil
}
