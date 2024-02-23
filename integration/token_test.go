package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
)

func TestGetToken(t *testing.T) {
	envs := getEnvs(t)

	client, err := bc.NewAuth(envs.TenantID, envs.ClientID, envs.ClientSecret, nil)
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

func TestGetTokenTimeout(t *testing.T) {

	tenantID := bc.GUID(os.Getenv("TENANT_ID"))
	clientID := bc.GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")

	client, err := bc.NewAuth(tenantID, clientID, clientSecret, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	_, err = client.GetToken(ctx)
	if err != nil {
		return
	}
	t.Error("did not time out correctly, token received")

}
