package testv2

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestAPIPageGetV2(t *testing.T) {

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

	ordersAPI, err := bc.NewAPIPage[salesOrder](client, "salesOrders")
	if err != nil {
		t.Fatalf("error at NewAPIPAge: %s", err)
	}

	ordersAPI.SetBaseExpand([]string{"salesOrderLines"})

	salesOrder, err := ordersAPI.Get(ctx, envs.SalesOrderRecordID, nil)
	if err != nil {
		t.Fatal(err)
	}

	b, _ := json.MarshalIndent(salesOrder, "", "  ")

	t.Logf("Item: %s", b)
}
