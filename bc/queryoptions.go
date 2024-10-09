package bc

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ListPageOptions build the Query
type ListPageOptions struct {
	Filter  string   // The filter expression. Combined with the BaseFilter.
	Expand  []string // The expandable fields. Added to the BaseExpand.
	OrderBy []string // The fields to order by, e.g. "field1 desc" or "field1". Ascending is default.
	Select  []string // The fields to return.
	Skip    int      // The number of records to skip. Do not use for pagination.
	Top     int      // The number of records to return. Do not use for pagination.
}

// BuildQueryParams combines the base filter/expand with the provided ListQueryOptions to return QueryParams
// for the request.
func (q *ListPageOptions) BuildQueryParams(baseFilter string, baseExpand []string) (QueryParams, error) {
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

	if len(q.OrderBy) > 0 {
		qp["$orderby"] = strings.Join(q.OrderBy, ",")
	}

	// Set $top if exists
	if q.Top != 0 {
		qp["$top"] = strconv.Itoa(q.Top)

		qp["$skip"] = strconv.Itoa(q.Skip)
	}

	return qp, nil
}
