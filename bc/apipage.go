package bc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// APIPage represents an API page in Business Central.
// It has the CRUD methods as well as a List method that returns
// a list of entities[T].
// Set a base filter with SetBaseFilter and an expand string with
// SetBaseExpand.
type APIPage[T Validator] struct {
	entitySetName string
	baseFilter    string
	client        *Client
	baseExpand    []string
}

// NewAPIPage creates an instance of an APIPage. It validates
// that the *Client is not nil and entitySetName is not empty. Call SetBaseExpand or SetBaseFilter
// on the APIPage to set the baseExpand/baseFilter.
func NewAPIPage[T Validator](client *Client, entitySetName string) (*APIPage[T], error) {
	if client == nil {
		return nil, errors.New("failed at NewAPIPage: client is nil")
	}

	if entitySetName == "" {
		return nil, errors.New("failed at NewAPIPage: entitySetName is empty")
	}
	return &APIPage[T]{
		entitySetName: entitySetName,
		client:        client,
	}, nil
}

// Sets the baseExpand slice. This will be added to all request expand expressions.
func (a *APIPage[T]) SetBaseExpand(expands []string) {
	a.baseExpand = slices.Concat(expands)
}

// Sets the base filter string. This will be added to all request filters.
func (a *APIPage[T]) SetBaseFilter(filter string) {
	a.baseFilter = filter
}

// Get makes a GET request to the endpoint and retrieves a single record T.
// Requires the ID and  takes an optional map of query options.
func (a *APIPage[T]) Get(ctx context.Context, id GUID, expand []string) (T, error) {
	var v T

	qp := QueryParams{}

	if len(expand) > 0 {
		qp["$expand"] = strings.Join(expand, ",")
	}

	opts := RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: a.entitySetName,
		RecordID:      id,
		QueryParams:   qp,
	}
	req, err := a.client.NewRequest(ctx, opts)
	if err != nil {
		return v, fmt.Errorf("failed to create Request: %w", err)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	v, err = Decode[T](res)
	var bcErr ServerError
	if err != nil {
		if errors.As(err, &bcErr) {
			return v, fmt.Errorf("error from server: %w", err)
		}
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil
}

// List makes a GET request to the endpoint and returns []T.
// It takes optional map of query options.
func (a *APIPage[T]) List(ctx context.Context, queryOpts ListQueryOptions) ([]T, error) {
	var v []T

	qp, err := queryOpts.BuildQueryParams(a.baseFilter, a.baseExpand)
	if err != nil {
		return v, fmt.Errorf("failed at BuildQueryParams: %w", err)
	}

	opts := RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: a.entitySetName,
		QueryParams:   qp,
	}
	req, err := a.client.NewRequest(ctx, opts)
	if err != nil {
		return v, fmt.Errorf("failed to create Request: %w", err)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	list, err := Decode[APIListResponse[T]](res)
	if err != nil {
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	v = *list.Value
	return v, nil
}

type APIListResponse[T Validator] struct {
	Value *[]T `json:"value"`
}

func (a APIListResponse[T]) Validate() error {
	// Make sure Value is initialized
	fmt.Printf("a.Value == %v\n", a.Value)
	if a.Value == nil {
		return errors.New("validation error: Value is empty")
	}
	// Validate first object
	if len(*a.Value) > 0 {
		err := (*a.Value)[0].Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
