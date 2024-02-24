package v2_test

import (
	"os"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

type envParams struct {
	Environment  string
	CompanyID    bc.GUID
	TenantID     bc.GUID
	ClientID     bc.GUID
	ClientSecret string
}

func getEnvs(t *testing.T) (params envParams) {
	t.Helper()

	params.TenantID = bc.GUID(os.Getenv("TENANT_ID"))
	params.ClientID = bc.GUID(os.Getenv("CLIENT_ID"))
	params.ClientSecret = os.Getenv("CLIENT_SECRET")
	params.Environment = os.Getenv("ENVIRONMENT")
	params.CompanyID = bc.GUID(os.Getenv("COMPANY_ID"))
	return
}
