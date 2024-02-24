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

// APIListResponse is the response body of a valid GET request that does not
// have a RecordID. The Value field has a slice of T.
type APIListResponse[T any] struct {
	Value []T `json:"value" validate:"required,dive"`
}

func (a APIListResponse[T]) Validate() error {
	return ValidateStruct(a)
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
	a.baseExpand = expands
}

// Adds a new string to the baseExpand slice. This will be added
// to all request expand expressions.
func (a *APIPage[T]) AddBaseExpand(expand string) {
	a.baseExpand = append(a.baseExpand, expand)
}

// Returns the BaseExpand.
func (a *APIPage[T]) BaseExpand() []string {
	return a.baseExpand
}

// Sets the base filter string. This will be added to all request filters.
func (a *APIPage[T]) SetBaseFilter(filter string) {
	a.baseFilter = filter
}

// Get makes a GET request to the endpoint and retrieves a single record T.
// Requires the ID and  takes an optional slice of expand strings.
func (a *APIPage[T]) Get(ctx context.Context, id GUID, expand []string) (T, error) {
	var v T

	qp := QueryParams{}

	// Add new expands to base
	expands := a.baseExpand
	if len(expand) > 0 {
		expands = slices.Concat(a.baseExpand, expand)
	}

	qp["$expand"] = strings.Join(expands, ",")

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
	var srvErr ServerError
	if err != nil {
		if errors.As(err, &srvErr) {
			return v, fmt.Errorf("error from server: %w", err)
		}
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil
}

// List makes a GET request to the endpoint and returns []T.
// It takes optional struct of query options.
func (a *APIPage[T]) List(ctx context.Context, queryOpts ListPageOptions) ([]T, error) {
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
	v = list.Value
	return v, nil
}

// Update makes a Patch request to the endpoint and returns T.
// It requires a body and a RecordID.
func (a *APIPage[T]) Update(ctx context.Context, id GUID, expand []string, body any) (T, error) {
	var v T

	qp := QueryParams{}

	// Add new expands to base
	expands := a.baseExpand
	if len(expand) > 0 {
		expands = slices.Concat(a.baseExpand, expand)
	}

	qp["$expand"] = strings.Join(expands, ",")

	opts := RequestOptions{
		Method:        http.MethodPatch,
		EntitySetName: a.entitySetName,
		RecordID:      id,
		QueryParams:   qp,
		Body:          body,
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
	var srvErr ServerError
	if err != nil {
		if errors.As(err, &srvErr) {
			return v, fmt.Errorf("error from server: %w", err)
		}
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil

}
