package v2_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

type extractID struct {
	ID string `json:"id"`
}

func (e extractID) Validate() error {
	if e.ID == "" {
		return errors.New("validation error: ID is empty")
	}
	return nil
}

func TestGetSalesOrderV2(t *testing.T) {
	envs := getEnvs(t)

	ctx := context.Background()
	authClient, err := bc.NewAuth(envs.TenantID, envs.ClientID, envs.ClientSecret, nil)
	if err != nil {
		t.Fatal(err)
	}

	client, err := bc.NewClient(bc.ClientConfig{
		TenantID:    envs.TenantID,
		CompanyID:   envs.CompanyID,
		Environment: "EL_DEV_0205",
		APIEndpoint: "v2.0",
	}, authClient)
	if err != nil {
		t.Fatal(err)
	}

	opts := bc.RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: "salesOrders",
		RecordID:      "afb373be-c787-ee11-817a-6045bd7b892b",
	}

	req, err := client.NewRequest(ctx, opts)
	// t.Logf("req: %s, %s, %s", req.URL.String(), req.Method, req.Header.Get("Authorization"))

	if err != nil {
		t.Fatalf("failed to create new request: %s", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed making request: %s", err)
	}

	collection, err := bc.Decode[extractID](res)
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("collection: %+v", collection)

}
