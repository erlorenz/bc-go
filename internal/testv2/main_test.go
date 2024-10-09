package testv2

import (
	"os"
	"testing"

	"github.com/erlorenz/bc-go/bc"
	"github.com/joho/godotenv"
)

var testConfig bc.ClientConfig

func TestMain(m *testing.M) {
	godotenv.Load("../../.env")

	testConfig.TenantID = os.Getenv("TENANT_ID")
	testConfig.ClientID = os.Getenv("CLIENT_ID")
	testConfig.ClientSecret = os.Getenv("CLIENT_SECRET")
	testConfig.CompanyID = os.Getenv("COMPANY_ID")
	testConfig.Environment = os.Getenv("ENVIRONMENT")
	testConfig.APIEndpoint = "v2.0"

	if testConfig.ClientSecret == "" {
		panic("Missing envs")
	}

	os.Exit(m.Run())

}
