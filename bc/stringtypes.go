package bc

import (
	"fmt"
	"unicode/utf8"
)

// GUID represents a Microsoft GUID and implements the Validator interface.
//
// Deprecated: GUID exists for compatibility and should not be used. Prefer
// using uuid.UUID to be compatible with all.
type GUID string

func (id GUID) Validate() error {
	if utf8.RuneCountInString(string(id)) != 36 {
		return fmt.Errorf("'%s' is not valid GUID", id)
	}
	return nil

}
