package bc_test

import (
	"fmt"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestBuildBaseURLCommon(t *testing.T) {
	config := bc.ClientConfig{
		TenantID:    validGUID,
		Environment: "TEST",
		APIEndpoint: "v2.0",
		CompanyID:   validGUID,
	}
	url, err := bc.BuildBaseURL(config)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("TestPath", func(t *testing.T) {
		want := fmt.Sprintf("/v2.0/%s/TEST/api/v2.0/companies(%s)", validGUID, validGUID)
		got := url.Path
		if want != got {
			t.Errorf("wanted %s, got %s", want, got)
		}
	})
	t.Run("TestHost", func(t *testing.T) {
		want := "api.businesscentral.dynamics.com"
		got := url.Host
		if want != got {
			t.Errorf("wanted %s, got %s", want, got)
		}
	})

	t.Run("TestScheme", func(t *testing.T) {
		want := "https"
		got := url.Scheme
		if want != got {
			t.Errorf("wanted %s, got %s", want, got)
		}
	})
}

func TestBuildBaseURLExt(t *testing.T) {
	config := bc.ClientConfig{
		TenantID:    validGUID,
		Environment: "TEST",
		APIEndpoint: "publisher/group/version",
		CompanyID:   validGUID,
	}
	url, err := bc.BuildBaseURL(config)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("TestPath", func(t *testing.T) {
		want := fmt.Sprintf("/v2.0/%s/TEST/api/publisher/group/version/companies(%s)", validGUID, validGUID)
		got := url.Path
		if want != got {
			t.Errorf("wanted %s, got %s", want, got)
		}
	})
}
