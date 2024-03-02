package bc_test

import (
	"strings"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestAPIPageExpand(t *testing.T) {

	client, _ := bc.NewClient(fakeConfig, fakeTokenGetter{})
	apiPage, _ := bc.NewAPIPage[fakeEntity](client, "fakeEntities")

	apiPage.AddBaseExpand("lines")
	want := "lines"
	got := strings.Join(apiPage.BaseExpand(), ",")

	if want != got {
		t.Errorf("add1: wanted %s, got %s", want, got)
	}

	apiPage.AddBaseExpand("lines")
	want = "lines,lines"
	got = strings.Join(apiPage.BaseExpand(), ",")

	if want != got {
		t.Errorf("add2: wanted %s, got %s", want, got)
	}

	apiPage.SetBaseExpand([]string{"lines", "customer"})
	want = "lines,customer"
	got = strings.Join(apiPage.BaseExpand(), ",")

	if want != got {
		t.Errorf("add3: wanted %s, got %s", want, got)
	}

}
