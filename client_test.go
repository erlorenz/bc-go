package bcgo

import (
	"context"
	"testing"
)

type mockTokenGetter struct{}

func (ac *mockTokenGetter) GetToken(context.Context) (AccessToken, error) {
	return "ACCESSTOKEN", nil
}

var config = ClientConfig{
	TenantID:     ValidGUID,
	Environment:  "Sandbox",
	APIPublisher: "publisher",
	APIGroup:     "group",
	APIVersion:   "1.0",
	CompanyID:    ValidGUID,
}

func TestNewClient(t *testing.T) {

	token := &mockTokenGetter{}

	client, err := NewClient(config, token)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("TestPubGroupVersion", func(t *testing.T) {

		want := "publisher/group/v1.0"
		got := client.PubGroup()
		if want != got {
			t.Errorf("expected %s, got %s", want, got)
		}
	})

}
