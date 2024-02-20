package integration

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestListSalesOrdersV2(t *testing.T) {

	ctx := context.Background()
	authClient := initAuthClient(t)

	// test get token
	token, err := authClient.GetToken(ctx)
	if err != nil {
		t.Fatalf("failed to get token %s", err)
	}
	t.Logf("token: %s", token)

	client := initClient(authClient, t)
	t.Logf("client %v", client)

	opts := bc.MakeRequestOptions{
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

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %s", err)
	}

	t.Logf("%s", string(body))

}
