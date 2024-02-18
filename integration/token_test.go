package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	bcgo "github.com/erlorenz/bc-go"
)

func TestGetToken(t *testing.T) {

	tenantID := bcgo.GUID(os.Getenv("TENANT_ID"))
	clientID := bcgo.GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")

	params := bcgo.AuthParams{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Logger:       slog.Default(),
	}

	client, err := bcgo.NewAuthClient(params)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*4)
	defer cancel()

	token, err := client.GetToken(ctx)
	if err != nil {
		t.Error(err)
	}
	t.Log("Access token retrieved:", token)
}
