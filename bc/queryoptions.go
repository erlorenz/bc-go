package bc

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const OrderAscending = "ASC"
const OrderDescending = "DESC"

// ListPageOptions are the different OData filters and expressions that are
// sent as query params in the request to an APIPage.
type ListPageOptions struct {
	Filter         string
	Expand         []string
	OrderBy        string
	OrderDirection string
	Skip           int
	Top            int
}

// BuildQueryParams combines the base filter/expand with the provided ListQueryOptions to return QueryParams
// for the request.
func (q ListPageOptions) BuildQueryParams(baseFilter string, baseExpand []string) (QueryParams, error) {

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
	validOrderDirections := []string{"", "ASC", "DESC"}

	if q.OrderBy != "" {
		if !slices.Contains(validOrderDirections, q.OrderDirection) {
			return nil, fmt.Errorf("invalid order direction '%s', must be ASC or DESC", q.OrderDirection)
		}
		dir := cmp.Or(q.OrderDirection, "ASC")
		qp["$orderby"] = fmt.Sprintf("%s %s", q.OrderBy, dir)
	}

	// Set $top if exists
	if q.Top != 0 {
		qp["$top"] = strconv.Itoa(q.Top)

		qp["$skip"] = strconv.Itoa(q.Skip)
	}

	return qp, nil
}
