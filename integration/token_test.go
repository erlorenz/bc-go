package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
)

func TestGetToken(t *testing.T) {

	tenantID := bc.GUID(os.Getenv("TENANT_ID"))
	clientID := bc.GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")

	params := bc.AuthParams{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Logger:       slog.Default(),
	}

	client, err := bc.NewAuthClient(params)
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
	t.Log("Access token retrieved, first 5:", token[:30])
}

func TestGetTokenTimeout(t *testing.T) {

	tenantID := bc.GUID(os.Getenv("TENANT_ID"))
	clientID := bc.GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")
	logger := slog.Default()

	params := bc.AuthParams{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Logger:       logger,
	}

	client, err := bc.NewAuthClient(params)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	_, err = client.GetToken(ctx)
	if err != nil {
		return
	}
	t.Error("did not time out correctly, token received")

}
