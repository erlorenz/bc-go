package bc

import (
	"fmt"
	"net/url"
	"unicode/utf8"
)

const EmptyString = ""

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
