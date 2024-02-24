package testv2

import (
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/erlorenz/bc-go/bc"
	"github.com/joho/godotenv"
)

type envParams struct {
	Environment        string  `validate:"required"`
	CompanyID          bc.GUID `validate:"required,uuid"`
	TenantID           bc.GUID `validate:"required,uuid"`
	ClientID           bc.GUID `validate:"required,uuid"`
	ClientSecret       string  `validate:"required"`
	SalesOrderRecordID bc.GUID `validate:"required"`
}

var envs envParams

var testClientV2 *bc.Client

func TestMain(m *testing.M) {

	godotenv.Load("../../.env")

	envs.TenantID = bc.GUID(os.Getenv("TENANT_ID"))
	envs.ClientID = bc.GUID(os.Getenv("CLIENT_ID"))
	envs.ClientSecret = os.Getenv("CLIENT_SECRET")
	envs.Environment = os.Getenv("ENVIRONMENT")
	envs.CompanyID = bc.GUID(os.Getenv("COMPANY_ID"))
	envs.SalesOrderRecordID = bc.GUID(os.Getenv("SALES_ORDER_RECORD_ID"))

	err := bc.ValidateStruct(envs)
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	auth, err := bc.NewAuth(envs.TenantID, envs.ClientID, envs.ClientSecret, logger)
	if err != nil {
		log.Fatal(err)
	}

	client, err := bc.NewClient(bc.ClientConfig{
		TenantID:    envs.TenantID,
		CompanyID:   envs.CompanyID,
		Environment: envs.Environment,
		APIEndpoint: "v2.0",
	}, auth, bc.WithLogger(logger))
	if err != nil {
		log.Fatal(err)
	}

	testClientV2 = client

	os.Exit(m.Run())
}
