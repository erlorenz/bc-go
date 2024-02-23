package bc_test

import (
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestNewAuthClient(t *testing.T) {

	type params struct {
		tenantID     bc.GUID
		clientID     bc.GUID
		clientSecret string
	}

	// false if error, true if success
	table := []struct {
		name   string
		params params
		want   bool
	}{
		{"GoodParams", params{validGUID, validGUID, "TEST"}, true},
		{"BadClientSecret", params{validGUID, validGUID, ""}, false},
		{"AllBad", params{"", "", ""}, false},
	}

	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			_, err := bc.NewAuth(v.params.tenantID, v.params.clientID, v.params.clientSecret, nil)
			want := v.want
			got := err == nil
			if want != got {
				if want == false {
					t.Errorf("expected no error, got %s", err)
					return
				}
				t.Error("expected error, got nil")
			}

		})
	}

}
