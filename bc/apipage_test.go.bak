package bc_test

import (
	"strings"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestAPIPageExpand(t *testing.T) {

	client, err := bc.NewClient(fakeConfig, bc.WithAuthClient(fakeTokenGetter{}))
	if err != nil {
		t.Fatalf("Did not expect error at newclient, got %s", err)
	}
	apiPage := bc.NewAPIPage[fakeEntity](client, "fakeEntities")

	apiPage.AddBaseExpand("lines")
	want := "lines"
	got := strings.Join(apiPage.BaseExpand, ",")

	if want != got {
		t.Errorf("add1: wanted %s, got %s", want, got)
	}

	apiPage.AddBaseExpand("lines")
	want = "lines,lines"
	got = strings.Join(apiPage.BaseExpand, ",")

	if want != got {
		t.Errorf("add2: wanted %s, got %s", want, got)
	}

	apiPage.BaseExpand = []string{"lines", "customer"}
	want = "lines,customer"
	got = strings.Join(apiPage.BaseExpand, ",")

	if want != got {
		t.Errorf("add3: wanted %s, got %s", want, got)
	}

}
