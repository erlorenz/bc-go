package bcgo

import (
	"errors"
	"fmt"
	"net/url"
	"unicode/utf8"
)

// GUID represents a Microsoft GUID and implements the Validator interface.
type GUID string

func (id GUID) Validate() error {
	if utf8.RuneCountInString(string(id)) != 36 {
		return fmt.Errorf("'%s' is not valid GUID", id)
	}
	return nil
}

// URL represents a URL string and implements the Validator interface
type URLString string

func (u URLString) Validate() error {
	_, err := url.ParseRequestURI(string(u))
	if err != nil {
		return fmt.Errorf("'%s' is not valid URL", u)
	}
	return nil
}

type NotEmptyString string

func (s NotEmptyString) Validate() error {
	if len(s) == 0 {
		return errors.New("is empty")
	}
	return nil
}
