package bc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type APIQuery[T any] struct {
	entitySetName string
	baseFilter    string
	client        *Client
}

func NewAPIQuery[T any](client *Client, entitySetName string) (*APIQuery[T], error) {
	if client == nil {
		return nil, errors.New("failed at NewAPIPage: client is nil")
	}

	if entitySetName == "" {
		return nil, errors.New("failed at NewAPIPage: entitySetName is empty")
	}
	return &APIQuery[T]{
		entitySetName: entitySetName,
		client:        client,
	}, nil
}

func (q *APIQuery[T]) SetBaseFilter(filter string) {
	q.baseFilter = filter
}

func (q *APIQuery[T]) List(ctx context.Context, filter string, orderby string, top int) ([]T, error) {
	var v []T

	filterStrings := []string{}
	// Add the baseFilter
	if q.baseFilter != "" {
		filterStrings = append(filterStrings, q.baseFilter)
	}

	// Add the filter and surround with "()" if there is a baseFilter
	if filter != "" {
		if q.baseFilter != "" {
			filter = fmt.Sprintf("(%s)", filter)
		}
		filterStrings = append(filterStrings, filter)
	}

	filter = strings.Join(filterStrings, " and ")

	qp := QueryParams{}
	if filter != "" {
		qp["$filter"] = filter
	}

	// Set $orderby if exists
	if orderby != "" {
		if orderby != "ASC" && orderby != "DESC" {
			return nil, fmt.Errorf("bad orderby format '%s', must be either 'DESC' or 'ASC'", orderby)
		}
		qp["$orderby"] = orderby
	}

	// Set $top if exists
	if top > 0 {
		qp["$top"] = strconv.Itoa(top)
	}

	opts := RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: q.entitySetName,
		QueryParams:   qp,
	}
	req, err := q.client.NewRequest(ctx, opts)
	if err != nil {
		return v, fmt.Errorf("failed to create Request: %w", err)
	}

	res, err := q.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	list, err := Decode[APIListResponse[T]](res)
	if err != nil {
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	v = list.Value
	return v, nil
}
