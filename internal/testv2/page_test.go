package testv2

import (
	"context"
	"fmt"
	"math/rand/v2"
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
	ctx := context.Background()

	client, err := bc.NewClient(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	itemsPage := bc.NewAPIPage[Item](client, "items")

	items, err := itemsPage.List(ctx, bc.ListOptions{
		Top: 2,
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 2 {
		t.Fatalf("wanted items length 2, got %d", len(items))
		t.Logf("items: %#v", items)
	}

	id := items[0].ID

	item, err := itemsPage.Get(ctx, id, bc.GetOptions{
		Select: []string{"id", "number"},
		Expand: []string{"itemCategory"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := id
	got := item.ID
	if want != got {
		t.Errorf("wanted %s, got %s", want, got)
	}

}

func TestAPIPage_New(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, err := bc.NewClient(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	num := fmt.Sprintf("TESTTEST%d", rand.IntN(100))

	body := map[string]any{
		"number":       num,
		"taxGroupCode": nil,
	}

	itemsPage := bc.NewAPIPage[Item](client, "items")
	item, err := itemsPage.Create(ctx, body, bc.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if item.Number != num {
		t.Fatalf("wanted %s, got %s", num, item.Number)
	}

	if err := itemsPage.Delete(ctx, item.ID); err != nil {
		t.Fatal(err)
	}

}
