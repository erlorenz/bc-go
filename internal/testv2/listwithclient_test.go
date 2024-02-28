package testv2

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestListSalesOrders(t *testing.T) {

	ctx := context.Background()

	opts := bc.RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: "salesOrders",
	}

	opts.QueryParams = bc.QueryParams{
		"$filter": "startswith(number, 'T')",
		"$expand": "salesOrderLines,customer",
	}

	req, err := testClientV2.NewRequest(ctx, opts)
	// t.Logf("req: %s, %s, %s", req.URL.String(), req.Method, req.Header.Get("Authorization"))

	if err != nil {
		t.Fatalf("failed to create new request: %s", err)
	}

	res, err := testClientV2.Do(req)
	if err != nil {
		t.Fatalf("failed making request: %s", err)
	}

	listResponse, err := bc.Decode[bc.APIListResponse[salesOrder]](res)
	if err != nil {

		var srvErr bc.APIError
		if errors.As(err, &srvErr) {
			t.Fatalf("server error: %+v", srvErr)
		}

		bytes, _ := io.ReadAll(res.Body)
		t.Logf("order: %+v", string(bytes))
		t.Fatalf("%s", err)
	}

	bytes, _ := json.MarshalIndent(listResponse.Value, "", "  ")
	t.Logf("orders: %+v", string(bytes))

}
