package integration

import (
	"log/slog"
	"os"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func initClient(authClient *bc.AuthClient, t *testing.T) *bc.Client {
	t.Helper()

	tenantID := bc.GUID(os.Getenv("TENANT_ID"))
	companyID := bc.GUID(os.Getenv("COMPANY_ID"))
	apiEndpoint := "v2.0"
	environment := os.Getenv("ENVIRONMENT")
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	config := bc.ClientConfig{
		TenantID:    tenantID,
		CompanyID:   companyID,
		APIEndpoint: apiEndpoint,
		Environment: environment,
	}

	client, err := bc.NewClient(config, authClient, bc.WithLogger(logger))
	if err != nil {
		t.Fatalf("err initializing client: %s", err)
	}

	return client
}

func initAuthClient(t *testing.T) *bc.AuthClient {
	t.Helper()

	tenantID := bc.GUID(os.Getenv("TENANT_ID"))
	clientID := bc.GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	authParams := bc.AuthParams{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Logger:       logger,
	}

	authClient, err := bc.NewAuthClient(authParams)
	if err != nil {
		t.Fatalf("failed initializing authclient: %s", err)
	}

	return authClient

}
