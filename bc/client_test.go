package bc_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
	"github.com/erlorenz/bc-go/internal/bctest"
	"github.com/google/uuid"
)

// FakeTokenGetter satisfies the TokenGetter interface.
// It returns a fake access token.
type fakeTokenGetter struct{}

// GetToken returns a fake access token and a nil error
func (ac fakeTokenGetter) GetToken(context.Context) (bc.AccessToken, error) {
	return bc.AccessToken("FAKEACCESSTOKEN"), nil
}

var fakeConfig = bc.ClientConfig{
	TenantID:     validGUID,
	Environment:  "Sandbox",
	APIEndpoint:  "publisher/group/1.0",
	CompanyID:    validGUID,
	ClientID:     validGUID,
	ClientSecret: "SECRET",
}

func TestNewClient(t *testing.T) {

	client, err := bc.NewClient(fakeConfig, bc.WithAuthClient(&fakeTokenGetter{}))
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
		got := client.IsV2()
		if want != got {
			t.Errorf("expected %t, got %t", want, got)
		}
	})

}

func TestNewClientCommon(t *testing.T) {

	// Make APIEndpoint v2.0
	config := fakeConfig
	config.APIEndpoint = "v2.0"

	client, err := bc.NewClient(config, bc.WithAuthClient(&fakeTokenGetter{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("APIEndpoint", func(t *testing.T) {

		want := "v2.0"
		got := client.APIEndpoint()
		if want != got {
			t.Logf("client: %+v", client)
			t.Errorf("expected %s, got %s", want, got)
		}
	})

	t.Run("IsCommon", func(t *testing.T) {

		want := true
		got := client.IsV2()
		if want != got {
			t.Logf("client: %+v", client)
			t.Errorf("expected %t, got %t", want, got)
		}
	})
	t.Run("LoggerDefault", func(t *testing.T) {

		want := true
		got := client.Logger() != nil
		if want != got {
			t.Logf("client: %+v", client)
			t.Errorf("expected %t, got %t", want, got)
		}
	})
	t.Run("HTTPClientDefault", func(t *testing.T) {

		want := true
		got := client.BaseClient() != nil
		if want != got {
			t.Logf("client: %+v", client)
			t.Errorf("expected %t, got %t", want, got)
		}
	})

}

func TestClientWithOptions(t *testing.T) {
	timeout := 60 * time.Minute
	httpClient := &http.Client{
		Timeout: timeout,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	client, err := bc.NewClient(fakeConfig, bc.WithAuthClient(&fakeTokenGetter{}), bc.WithHTTPClient(httpClient), bc.WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("WithHTTPClient", func(t *testing.T) {
		want := timeout
		got := client.BaseClient().Timeout

		if got != want {
			t.Logf("client: %+v", client)
			t.Errorf("want %v, got %v", want, got)
		}
	})
	t.Run("WithLogger", func(t *testing.T) {
		want := logger
		got := client.Logger()
		if got != want {
			t.Logf("client: %+v", client)
			t.Errorf("want %v, got %v", want, got)
		}
	})

}

func TestConfig(t *testing.T) {

	validConfig := bc.ClientConfig{
		TenantID:     uuid.NewString(),
		CompanyID:    uuid.NewString(),
		Environment:  "TEST",
		APIEndpoint:  "v2.0",
		ClientID:     uuid.NewString(),
		ClientSecret: "SECRET",
	}

	emptyConfig := bc.ClientConfig{}

	missingCompanyID := validConfig
	missingCompanyID.CompanyID = bctest.EmptyString

	invalidGUID := validConfig
	invalidGUID.TenantID = "NOT A GUID"

	type testCase struct {
		name       string
		config     bc.ClientConfig
		shouldPass bool
	}

	table := []testCase{
		{"valid", validConfig, true},
		{"empty", emptyConfig, false},
		{"missing CompanyID", missingCompanyID, false},
		{"invalid GUID", invalidGUID, false},
	}

	for _, test := range table {
		_, err := bc.NewClient(test.config, bc.WithAuthClient(fakeTokenGetter{}))
		passed := err == nil
		if test.shouldPass != passed {
			t.Errorf("%s: wanted %t, got %t: %s", test.name, test.shouldPass, passed, err)
		}

	}

}
