package bcgo

import (
	"fmt"
	"net/url"
)

// GUID represents a Microsoft GUID and implements the Validator interface.
type GUID string

func (id GUID) Validate() error {
	if len(id) != 36 {
		return fmt.Errorf("%s is not valid GUID", id)
	}
	return nil
}

// URL represents a URL string and implements the Validator interface
type URLString string

func (u URLString) Validate() error {
	if _, err := url.Parse(string(u)); err != nil {
		return fmt.Errorf("%s is not valid URL", u)
	}
	return nil
}

type Validator interface {
	Validate() error
}
