package testv2

import (
	"context"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestAPIPage_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		pan := recover()
		if pan == nil {
			t.Fatalf("expected panic, got nil")
		}
	}()

	client, err := bc.NewClient(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	bc.NewAPIPage[Item](client, "")

}

func TestAPIPage_Get(t *testing.T) {
	t.Parallel()

	client, err := bc.NewClient(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	itemsPage := bc.NewAPIPage[Item](client, "items")

	items, err := itemsPage.List(context.Background(), bc.ListPageOptions{
		Top: 2,
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 2 {
		t.Fatalf("wanted items length 2, got %d", len(items))
		t.Logf("items: %#v", items)
	}

}
