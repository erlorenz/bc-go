package bc

import (
	"context"
	"testing"
)

type fakeTokenGetter struct{}

func (ac *fakeTokenGetter) GetToken(context.Context) (AccessToken, error) {
	return "ACCESSTOKEN", nil
}

var fakeToken = &fakeTokenGetter{}

var fakeConfig = ClientConfig{
	TenantID:    validGUID,
	Environment: "Sandbox",
	APIEndpoint: "publisher/group/1.0",
	CompanyID:   validGUID,
}

func TestNewClient(t *testing.T) {

	client, err := NewClient(fakeConfig, fakeToken)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("APIEndpoint", func(t *testing.T) {

		want := "publisher/group/1.0"
		got := client.APIEndpoint()
		if want != got {
			t.Errorf("expected %s, got %s", want, got)
		}
	})

	t.Run("IsCommon", func(t *testing.T) {

		want := false
		got := client.IsCommon()
		if want != got {
			t.Errorf("expected %t, got %t", want, got)
		}
	})

}

func TestNewClientCommon(t *testing.T) {

	// Make APIEndpoint v2.0
	config := fakeConfig
	config.APIEndpoint = "v2.0"

	client, err := NewClient(config, fakeToken)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("APIEndpoint", func(t *testing.T) {

		want := "v2.0"
		got := client.APIEndpoint()
		if want != got {
			t.Errorf("expected %s, got %s", want, got)
		}
	})

	t.Run("IsCommon", func(t *testing.T) {

		want := true
		got := client.IsCommon()
		if want != got {
			t.Errorf("expected %t, got %t", want, got)
		}
	})

}
func TestErrorClient(t *testing.T) {

}
