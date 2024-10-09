package bc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// APIQuery interacts with a BC API Query.
type APIQuery[T any] struct {
	entitySetName string
	client        *Client
	BaseFilter    string
	BaseSelect    []string
}

// NewAPIQuery returns an [APIQuery]. It panics if missing a client or entitySetName.
func NewAPIQuery[T any](client *Client, entitySetName string) *APIQuery[T] {
	if client == nil {
		panic("create API query: client is nil")
	}
	if entitySetName == "" {
		panic("create API query: entitySetName is empty")
	}

	return &APIQuery[T]{
		entitySetName: entitySetName,
		client:        client,
	}
}

// List
func (q *APIQuery[T]) List(ctx context.Context, opts ListPageOptions) ([]T, error) {
	var v []T

	filterStrings := []string{}
	// Add the baseFilter
	if q.BaseFilter != "" {
		filterStrings = append(filterStrings, q.BaseFilter)
	}

	// Add the filter and surround with "()" if there is a baseFilter
	if opts.Filter != "" {
		if q.BaseFilter != "" {
			opts.Filter = fmt.Sprintf("(%s)", opts.Filter)
		}
		filterStrings = append(filterStrings, opts.Filter)
	}

	filter := strings.Join(filterStrings, " and ")

	qp := QueryParams{}

	if filter != "" {
		qp["$filter"] = filter
	}

	// Set $top if exists
	if opts.Top > 0 {
		qp["$top"] = strconv.Itoa(opts.Top)
	}

	ropts := RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: q.entitySetName,
		QueryParams:   qp,
	}
	req, err := q.client.NewRequest(ctx, ropts)
	if err != nil {
		return v, fmt.Errorf("failed to create Request: %w", err)
	}

	res, err := q.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	list, err := Decode[APIListResponse[T]](res)
	if err != nil {
		var srvErr APIError
		if errors.As(err, &srvErr) {
			q.client.logger.Debug("API server returned error response.", "error", srvErr)
			return v, fmt.Errorf("error from BC API: %w", srvErr)
		}

		q.client.logger.Debug("Failed to decode response.", "error", err)
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	v = list.Value
	return v, nil
}
