package testv2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/erlorenz/bc-go/bc"
	"github.com/erlorenz/bc-go/internal/bctest"
	"github.com/google/uuid"
)

type Item struct {
	ID     uuid.UUID `json:"id"`
	Number string    `json:"number"`
}

type ItemList struct {
	Value []Item `json:"value"`
}

func (i Item) Validate() error {
	var errs []string

	if i.Number == "" {
		errs = append(errs, "number is empty")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (il ItemList) Validate() error {
	for index, item := range il.Value {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", index, err)
		}

	}
	return nil
}

func TestV2Client_GetItems(t *testing.T) {
	t.Parallel()
	client, err := bc.NewClient(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	req, err := client.NewRequest(context.Background(), bc.RequestOptions{
		Method:        "GET",
		EntitySetName: "items",
		QueryParams: bc.QueryParams{
			"$select": "id,number",
			"$top":    "5",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = bc.Decode[ItemList](res)
	if err != nil {
		t.Fatal(err)
	}

}

func TestV2Client_GetItemsAPIError(t *testing.T) {
	t.Parallel()
	client, err := bc.NewClient(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	req, err := client.NewRequest(context.Background(), bc.RequestOptions{
		Method:        "GET",
		EntitySetName: "itemsa",
		QueryParams: bc.QueryParams{
			"$select": "id,number",
			"$top":    "5",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = bc.Decode[ItemList](res)
	if err == nil {
		t.Fatal(err)
	}

	var apiError bc.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected API error, got %s", err)
	}

	// want := http.StatusNotFound
	want := 404
	got := apiError.StatusCode
	if want != got {
		t.Fatalf("expected status %d, got %d", want, got)
	}

}

func TestV2Client_GetItemsNetworkError(t *testing.T) {
	t.Parallel()

	mhc := &http.Client{Transport: bctest.MockTransport{
		Error: errors.New("network_error"),
	}}

	client, err := bc.NewClient(testConfig, bc.WithHTTPClient(mhc))
	if err != nil {
		t.Fatal(err)
	}

	req, err := client.NewRequest(context.Background(), bc.RequestOptions{
		Method:        "GET",
		EntitySetName: "items",
		QueryParams:   bc.QueryParams{"$top": "5"},
	})

	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

}
