package bc

import (
	"testing"
)

func TestBuildQueryParams(t *testing.T) {

	filter := "number eq 'XXXX'"
	expands := []string{"salesLines", "customer"}

	opts := ListQueryOptions{
		Filter:  filter,
		Expand:  expands,
		Top:     5,
		Orderby: "ASC",
	}

	qp, err := opts.BuildQueryParams("", nil)
	if err != nil {
		t.Fatalf("failed at BuildQueryParams: %s", err)
	}

	if qp["$filter"] != filter {
		t.Errorf(`wrong filter: expected "%s", got "%s"`, filter, qp["$filter"])
	}

	if qp["$expand"] != "salesLines,customer" {
		t.Errorf(`wrong expand: expected "salesLines,customer", got "%s"`, qp["$expand"])
	}

	if qp["$top"] != "5" {
		t.Errorf(`wrong top: expected "5", got "%s"`, qp["$top"])
	}

	if qp["$orderby"] != "ASC" {
		t.Errorf(`wrong top: expected "ASC", got "%s"`, qp["$orderby"])
	}

}

func TestBuildQueryParamsWithBase(t *testing.T) {

	filter := "number eq 'XXXX'"
	expands := []string{"salesLines", "customer"}

	opts := ListQueryOptions{
		Filter:  filter,
		Expand:  expands,
		Top:     5,
		Orderby: "ASC",
	}

	baseFilter := "documentType eq 'Order'"
	baseExpand := []string{"dimensionSetLines"}

	expectedFilter := "documentType eq 'Order' and (number eq 'XXXX')"
	expectedExpand := "dimensionSetLines,salesLines,customer"

	qp, err := opts.BuildQueryParams(baseFilter, baseExpand)
	if err != nil {
		t.Fatalf("failed at BuildQueryParams: %s", err)
	}

	if qp["$filter"] != expectedFilter {
		t.Errorf(`wrong filter: expected "%s", got "%s"`, expectedFilter, qp["$filter"])
	}

	if qp["$expand"] != expectedExpand {
		t.Errorf(`wrong expand: expected "%s", got "%s"`, expectedExpand, qp["$expand"])
	}

	if qp["$top"] != "5" {
		t.Errorf(`wrong top: expected "5", got "%s"`, qp["$top"])
	}

	if qp["$orderby"] != "ASC" {
		t.Errorf(`wrong top: expected "ASC", got "%s"`, qp["$orderby"])
	}

}
