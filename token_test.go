package bcgo

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"time"

	"github.com/joho/godotenv"
)

const ValidGUID GUID = "ecd89ac3-1f77-48db-b42c-2640119cc69a"

func TestNewAuthClient(t *testing.T) {

	goodParams := AuthParams{
		TenantID:     ValidGUID,
		ClientID:     ValidGUID,
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

func TestGetTokenTimeout(t *testing.T) {

	godotenv.Load(".env")
	tenantID := GUID(os.Getenv("TENANT_ID"))
	clientID := GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")
	logger := slog.Default()

	params := AuthParams{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Logger:       logger,
	}

	client, err := NewAuthClient(params)
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
