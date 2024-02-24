package v2_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/erlorenz/bc-go/bc"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {

	godotenv.Load("../.env")

	tenantID := bc.GUID(os.Getenv("TENANT_ID"))
	clientID := bc.GUID(os.Getenv("CLIENT_ID"))
	clientSecret := os.Getenv("CLIENT_SECRET")

	if tenantID == "" || clientID == "" || clientSecret == "" {
		fmt.Println("Missing env variables.")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
