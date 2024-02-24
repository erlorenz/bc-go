package testv2

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestAPIPagePatchV2(t *testing.T) {

	ctx := context.Background()

	ordersAPI, err := bc.NewAPIPage[salesOrder](testClientV2, "salesOrders")
	if err != nil {
		t.Fatalf("error at NewAPIPAge: %s", err)
	}

	ordersAPI.SetBaseExpand([]string{"salesOrderLines"})

	updateDate, _ := bc.ParseDate("2024-02-15")
	body := struct {
		OrderDate bc.Date `json:"orderDate"`
	}{updateDate}

	salesOrder, err := ordersAPI.Update(ctx, envs.SalesOrderRecordID, nil, body)
	if err != nil {
		t.Fatal(err)
	}

	if salesOrder.OrderDate.String() != updateDate.String() {
		t.Errorf("dates were not the same, wanted %s, got %s", updateDate.String(), salesOrder.OrderDate.String())
	}

	b, _ := json.MarshalIndent(salesOrder, "", "  ")
	t.Logf("Item: %s", b)

}
