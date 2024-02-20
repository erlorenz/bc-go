package bc

import (
	"log/slog"
	"testing"
)

const validGUID GUID = "ecd89ac3-1f77-48db-b42c-2640119cc69a"

func TestNewAuthClient(t *testing.T) {

	goodParams := AuthParams{
		TenantID:     validGUID,
		ClientID:     validGUID,
		ClientSecret: "TESTSECRET",
		Logger:       slog.Default(),
	}

	badGUIDParams := goodParams
	badGUIDParams.ClientID = "BAD"

	badSecretParams := goodParams
	badSecretParams.ClientSecret = ""

	multipleBadParams := badGUIDParams
	multipleBadParams.ClientSecret = ""
	multipleBadParams.TenantID = ""

	testTable := map[string]struct {
		params    AuthParams
		wantError bool
	}{
		"GoodParams":  {goodParams, false},
		"BadGUID":     {badGUIDParams, true},
		"BadSecret":   {badSecretParams, true},
		"MultipleBad": {multipleBadParams, true},
	}

	for k, v := range testTable {
		t.Run(k, func(t *testing.T) {
			_, err := NewAuthClient(v.params)

			gotError := err != nil
			if v.wantError != gotError {
				if v.wantError == false {
					t.Errorf("expected no error, got %s", err)
					return
				}
				t.Error("expected error, got nil")
			}

		})
	}

}
