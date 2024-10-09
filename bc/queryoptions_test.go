package bc

import (
	"testing"
)

func TestBuildQueryParams_List(t *testing.T) {

	filter := "number eq 'XXXX'"
	expands := []string{"salesLines", "customer"}

	opts := ListOptions{
		Filter:  filter,
		Expand:  expands,
		Top:     5,
		OrderBy: []string{"number asc"},
	}

	qp := opts.BuildQueryParams("", nil)

	if qp["$filter"] != filter {
		t.Errorf(`wrong filter: expected "%s", got "%s"`, filter, qp["$filter"])
	}

	if qp["$expand"] != "salesLines,customer" {
		t.Errorf(`wrong expand: expected "salesLines,customer", got "%s"`, qp["$expand"])
	}

	if qp["$top"] != "5" {
		t.Errorf(`wrong top: expected "5", got "%s"`, qp["$top"])
	}

	if qp["$skip"] != "0" {
		t.Errorf(`wrong skip: expected "0", got "%s"`, qp["$skip"])
	}

	if qp["$orderby"] != "number asc" {
		t.Errorf(`wrong orderby: expected "number asc", got "%s"`, qp["$orderby"])
	}

	opts = ListOptions{
		Top:     10,
		Skip:    20,
		OrderBy: []string{"number desc"},
	}
	qp = opts.BuildQueryParams("", nil)

	if qp["$skip"] != "20" {
		t.Errorf(`wrong skip: expected "20", got "%s"`, qp["$skip"])
	}

	if qp["$orderby"] != "number desc" {
		t.Errorf(`wrong orderby: expected "number desc", got "%s"`, qp["$orderby"])
	}

	opts.OrderBy = []string{"number"}

}

func TestBuildQueryParamsWithBase_List(t *testing.T) {

	filter := "number eq 'XXXX'"
	expands := []string{"salesLines", "customer"}

	opts := ListOptions{
		Filter:  filter,
		Expand:  expands,
		Top:     5,
		OrderBy: []string{"number asc"},
	}

	baseFilter := "documentType eq 'Order'"
	baseExpand := []string{"dimensionSetLines"}

	expectedFilter := "documentType eq 'Order' and (number eq 'XXXX')"
	expectedExpand := "dimensionSetLines,salesLines,customer"

	qp := opts.BuildQueryParams(baseFilter, baseExpand)

	if qp["$filter"] != expectedFilter {
		t.Errorf(`wrong filter: expected "%s", got "%s"`, expectedFilter, qp["$filter"])
	}

	if qp["$expand"] != expectedExpand {
		t.Errorf(`wrong expand: expected "%s", got "%s"`, expectedExpand, qp["$expand"])
	}

	if qp["$top"] != "5" {
		t.Errorf(`wrong top: expected "5", got "%s"`, qp["$top"])
	}

	if qp["$orderby"] != "number asc" {
		t.Errorf(`wrong orderby: expected "number asc", got "%s"`, qp["$orderby"])
	}

}

func TestBuildQueryParamsWithBase_Get(t *testing.T) {

	expands := []string{"salesLines", "customer"}
	opts := GetOptions{
		Expand: expands,
		Select: []string{"id", "number"},
	}

	baseExpand := []string{"dimensionSetLines"}

	expectedExpand := "dimensionSetLines,salesLines,customer"

	qp := opts.BuildQueryParams(baseExpand)

	if qp["$expand"] != expectedExpand {
		t.Errorf(`wrong expand: expected "%s", got "%s"`, expectedExpand, qp["$expand"])
	}

}
