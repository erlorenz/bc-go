package bc

import (
	"testing"
)

func TestURLStringValidate(t *testing.T) {

	goodURLs := []URLString{"https://www.test.com", "http://www.test.com/test234", "http://www.test.com/test244?sfjf=jskfj"}

	for _, v := range goodURLs {
		t.Run(string(v), func(t *testing.T) {

			err := v.Validate()
			if err != nil {
				t.Errorf("%s returned error", v)
			}
		})
	}

	badURLs := []URLString{"www.test.com", "", "notaurl"}

	for _, v := range badURLs {
		t.Run(string(v), func(t *testing.T) {

			err := v.Validate()
			if err == nil {
				t.Errorf("%s did not return error", v)
			}
		})
	}

}

func TestNotEmptyString(t *testing.T) {
	empty := NotEmptyString("")
	t.Run("Empty", func(t *testing.T) {
		err := empty.Validate()
		if err == nil {
			t.Errorf("did not error on empty string")
		}
	})

	notempty := NotEmptyString("NOTEMPTY")
	t.Run("NotEmpty", func(t *testing.T) {
		err := notempty.Validate()
		if err != nil {
			t.Error(err)
		}
	})

}
func TestGUID(t *testing.T) {
	empty := GUID("")
	t.Run("Empty", func(t *testing.T) {
		err := empty.Validate()
		if err == nil {
			t.Errorf("did not error on empty string")
		}
	})

	notGUID := GUID("NOTGUID")
	t.Run("NotGUID", func(t *testing.T) {
		err := notGUID.Validate()
		if err == nil {
			t.Error("did not error on empty string")
		}
	})

	isGUID := GUID("a2adda3d-e909-4b16-bb7d-da2af5e3c364")
	t.Run("IsGUID", func(t *testing.T) {
		err := isGUID.Validate()
		if err != nil {
			t.Error(err)
		}
	})

}
