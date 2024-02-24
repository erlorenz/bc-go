package v2_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
)

type salesOrder struct {
	ID                      string    `json:"id"`
	Number                  string    `json:"number"`
	OrderDate               string    `json:"orderDate"`
	CustomerID              string    `json:"customerId"`
	CustomerNumber          string    `json:"customerNumber"`
	CustomerName            string    `json:"customerName"`
	TotalAmountExcludingTax float64   `json:"totalAmountExcludingTax"`
	FullyShipped            bool      `json:"fullyShipped"`
	LastModifiedDateTime    time.Time `json:"lastModifiedDateTime"`
}

func (i salesOrder) Validate() error {
	if i.ID == "" {
		return errors.New("ID is empty")
	}
	return nil
}

func TestAPIPageGetV2(t *testing.T) {
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

	salesOrdersAPIPage, err := bc.NewAPIPage[salesOrder](client, "salesOrders")
	if err != nil {
		t.Fatalf("error at NewAPIPAge: %s", err)
	}
	salesOrder, err := salesOrdersAPIPage.Get(ctx, bc.GUID("e4ba78f4-f1c5-ee11-9078-000d3a9dc4b3"), nil)
	if err != nil {
		t.Fatal(err)
	}

	b, _ := json.MarshalIndent(salesOrder, "", "  ")

	t.Logf("Item: %s", b)
}

func TestAPIPageListV2(t *testing.T) {
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

	salesOrdersAPIPage, err := bc.NewAPIPage[salesOrder](client, "salesOrders")
	if err != nil {
		t.Fatalf("error at NewAPIPAge: %s", err)
	}
	salesOrders, err := salesOrdersAPIPage.List(ctx, bc.ListQueryOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Log only the IDs
	onlyIDs := mapslice(salesOrders, func(order salesOrder) extractID {
		return extractID{ID: order.ID}
	})
	b, _ := json.MarshalIndent(onlyIDs[:5], "", "  ")
	t.Logf("Items: %s", b)
}
