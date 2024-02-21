package bc

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ListQueryOptions are the different OData filters and expressions that are
// sent as query params in the request.
type ListQueryOptions struct {
	Filter  string
	Expand  []string
	Orderby string
	Top     int
}

// BuildQueryParams combines the base filter/expand with the provided ListQueryOptions to return QueryParams
// for the request.
func (q ListQueryOptions) BuildQueryParams(baseFilter string, baseExpand []string) (QueryParams, error) {

	// Filter should be in format "<baseFilter> and (<extrafilter>)"
	// Only supports adding to the base filter --don't use base filter if you need an "or"
	filterStrings := []string{}

	// Add the baseFilter
	if baseFilter != "" {
		filterStrings = append(filterStrings, baseFilter)
	}

	// Add the filter and surround with "()" if there is a baseFilter
	if q.Filter != "" {
		filter := q.Filter
		if baseFilter != "" {
			filter = fmt.Sprintf("(%s)", filter)
		}
		filterStrings = append(filterStrings, filter)
	}

	filter := strings.Join(filterStrings, " and ")

	// Expand should be comma separated
	expandSlice := slices.Concat(baseExpand, q.Expand)
	expand := strings.Join(expandSlice, ",")

	qp := QueryParams{}

	// Set $filter if exists
	if filter != "" {
		qp["$filter"] = filter
	}

	// Set $expand if exists
	if expand != "" {
		qp["$expand"] = expand
	}

	// Set $orderby if exists
	if q.Orderby != "" {
		if q.Orderby != "ASC" && q.Orderby != "DESC" {
			return nil, fmt.Errorf("bad orderby format '%s', must be either 'DESC' or 'ASC'", q.Orderby)
		}
		qp["$orderby"] = q.Orderby
	}

	// Set $top if exists
	if q.Top != 0 {
		qp["$top"] = strconv.Itoa(q.Top)
	}

	return qp, nil
}
