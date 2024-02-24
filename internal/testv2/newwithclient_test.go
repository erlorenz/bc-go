package testv2

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
)

func TestNewSalesOrderV2(t *testing.T) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	opts := bc.RequestOptions{
		Method:        http.MethodPost,
		EntitySetName: "salesOrders",
	}

	opts.QueryParams = bc.QueryParams{
		"$expand": "salesOrderLines($filter=lineType ne 'Comment')",
	}

	randomNumber := "T00" + strconv.Itoa(rand.Intn(10_000))

	opts.Body = newSalesOrderRequest{
		Number:         randomNumber,
		CustomerNumber: "TEST01",
	}

	req, err := testClientV2.NewRequest(ctx, opts)
	// t.Logf("req: %s, %s", req.URL.String(), req.Method)

	if err != nil {
		t.Fatalf("failed to create new request: %s", err)
	}

	res, err := testClientV2.Do(req)
	if err != nil {
		t.Fatalf("failed making request: %s", err)
	}

	order, err := bc.Decode[salesOrder](res)
	if err != nil {

		var srvErr bc.ServerError
		if errors.As(err, &srvErr) {
			t.Fatalf("server error: %+v", srvErr)
		}

		bytes, _ := io.ReadAll(res.Body)
		t.Logf("order: %+v", string(bytes))
		t.Fatalf("%s", err)
	}

	bytes, _ := json.MarshalIndent(order, "", "  ")
	t.Logf("order: %+v", string(bytes))

}
