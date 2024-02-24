package v2_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestListSalesOrdersV2(t *testing.T) {
	envs := getEnvs(t)

	ctx := context.Background()
	authClient, err := bc.NewAuth(envs.TenantID, envs.ClientID, envs.ClientSecret, nil)
	if err != nil {
		t.Fatal(err)
	}

	// test get token
	token, err := authClient.GetToken(ctx)
	if err != nil {
		t.Fatalf("failed to get token %s", err)
	}
	t.Logf("token: %s", token)

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

	collection, err := bc.Decode[bc.APIListResponse[extractID]](res)
	if err != nil {
		// Just do a ReadAll dump
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed reading body: %s", err)
		}
		t.Logf("%s", string(body))

	}

	t.Logf("collection: %+v", collection)

}
