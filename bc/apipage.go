package bc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// APIPage represents an API page in Business Central.
// It has the CRUD methods as well as a List method that returns
// a list of entities[T].
// Set a base filter with SetBaseFilter and an expand string with
// SetBaseExpand.
type APIPage[T Validator] struct {
	entitySetName string
	client        *Client
	BaseFilter    string
	BaseExpand    []string
}

// APIListResponse is the response body of a valid GET request that does not
// have a RecordID. The Value field has a slice of T.
type APIListResponse[T any] struct {
	Value []T `json:"value" validate:"required,dive"`
}

// Validate implements the Validator interface. It validates
// each row.
func (a APIListResponse[T]) Validate() error {
	return ValidateStruct(a)
}

// NewAPIPage creates an instance of an APIPage. It panics if client or entitySetName are empty.
// Call SetBaseExpand or SetBaseFilter to set filters/expand for all requests.
func NewAPIPage[T Validator](client *Client, entitySetName string) *APIPage[T] {
	if client == nil {
		panic("create API page: client is nil")
	}

	if entitySetName == "" {
		panic("create API page: entitySetName is empty")
	}
	return &APIPage[T]{
		entitySetName: entitySetName,
		client:        client,
	}
}

// Adds a new string to the baseExpand slice. This will be added
// to all request expand expressions.
func (a *APIPage[T]) AddBaseExpand(expand string) {
	a.BaseExpand = append(a.BaseExpand, expand)
}

// Get makes a GET request to the endpoint and retrieves a single record T.
// Requires the ID and  takes an optional slice of expand strings.
func (a *APIPage[T]) Get(ctx context.Context, id uuid.UUID, opts GetOptions) (T, error) {
	var v T

	qp := QueryParams{}

	// Add new expands to base
	expands := a.BaseExpand
	if len(opts.Expand) > 0 {
		expands = slices.Concat(a.BaseExpand, opts.Expand)
	}

	qp["$expand"] = strings.Join(expands, ",")

	reqOpts := RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: a.entitySetName,
		RecordID:      id,
		QueryParams:   qp,
	}
	req, err := a.client.NewRequest(ctx, reqOpts)
	if err != nil {
		return v, fmt.Errorf("failed to create Request: %w", err)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	v, err = Decode[T](res)
	if err != nil {
		var srvErr APIError
		if errors.As(err, &srvErr) {
			a.client.logger.Debug("API server returned error response.", "error", srvErr)
			return v, fmt.Errorf("error from BC API: %w", srvErr)
		}

		a.client.logger.Debug("Failed to decode response.", "error", err)
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil
}

// List makes a GET request to the endpoint and returns []T.
// It takes optional struct of query options.
func (a *APIPage[T]) List(ctx context.Context, queryOpts ListOptions) ([]T, error) {
	var v []T

	qp := queryOpts.BuildQueryParams(a.BaseFilter, a.BaseExpand)

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
		var srvErr APIError
		if errors.As(err, &srvErr) {
			a.client.logger.Debug("API server returned error response.", "error", srvErr)
			return v, fmt.Errorf("error from BC API: %w", srvErr)
		}

		a.client.logger.Debug("Unable to decode response.", "error", err)
		return v, fmt.Errorf("decode response: %w", err)
	}
	v = list.Value
	return v, nil
}

// Update makes a Patch request to the endpoint and returns T.
// It requires a body and a RecordID.
func (a *APIPage[T]) Update(ctx context.Context, id uuid.UUID, expand []string, body any) (T, error) {
	var v T

	qp := QueryParams{}

	// Add new expands to base
	expands := a.BaseExpand
	if len(expand) > 0 {
		expands = slices.Concat(a.BaseExpand, expand)
	}

	if len(expands) > 0 {
		qp["$expand"] = strings.Join(expands, ",")
	}

	a.client.logger.Debug("Query params initialized.", "expand", qp["$expand"])

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

	a.client.logger.Debug("Sending request...", "url", req.URL.String(), "method", req.Method)

	res, err := a.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	v, err = Decode[T](res)
	if err != nil {
		var srvErr APIError
		if errors.As(err, &srvErr) {
			a.client.logger.Debug("API server returned error response.", "error", srvErr)
			return v, fmt.Errorf("error from BC API: %w", srvErr)
		}

		a.client.logger.Debug("Failed to decode response.", "error", err)
		return v, fmt.Errorf("failed to decode response: %w", err)
	}

	a.client.logger.Debug(fmt.Sprintf("Successfully created %T record.", v), "record", fmt.Sprintf("%#v", v))
	return v, nil
}

// Create makes a POST request to the endpoint and returns T.
// It requires a body.
func (a *APIPage[T]) Create(ctx context.Context, body any, opts GetOptions) (T, error) {
	var v T

	qp := QueryParams{}

	// Add new expands to base
	expands := a.BaseExpand
	if len(opts.Expand) > 0 {
		expands = slices.Concat(a.BaseExpand, opts.Expand)
	}

	if len(expands) > 0 {
		qp["$expand"] = strings.Join(expands, ",")
	}
	a.client.logger.Debug("Query params initialized.", "expand", qp["$expand"])

	reqOpts := RequestOptions{
		Method:        http.MethodPost,
		EntitySetName: a.entitySetName,
		QueryParams:   qp,
		Body:          body,
	}
	req, err := a.client.NewRequest(ctx, reqOpts)
	if err != nil {
		return v, fmt.Errorf("failed to create Request: %w", err)
	}

	a.client.logger.Debug("Request initialized.", "url", req.URL.String(), "method", req.Method)

	res, err := a.client.Do(req)
	if err != nil {
		return v, fmt.Errorf("failed during request: %w", err)
	}

	v, err = Decode[T](res)
	if err != nil {
		var srvErr APIError
		if errors.As(err, &srvErr) {
			a.client.logger.Debug("API server returned error response.", "error", srvErr)
			return v, fmt.Errorf("error from BC API: %w", srvErr)
		}

		a.client.logger.Debug("Failed to decode response.", "error", err)
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	a.client.logger.Debug(fmt.Sprintf("Successfully created %T record.", v), "record", fmt.Sprintf("%#v", v))
	return v, nil

}

// Delete makes a DELETE request to the endpoint and returns a string message.
// It requires a RecordID.
func (a *APIPage[T]) Delete(ctx context.Context, id uuid.UUID) error {
	opts := RequestOptions{
		Method:        http.MethodDelete,
		EntitySetName: a.entitySetName,
		RecordID:      id,
	}
	req, err := a.client.NewRequest(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create Request: %w", err)
	}

	a.client.logger.Debug("Sending request...", "url", req.URL.String(), "method", req.Method)

	res, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed during request: %w", err)
	}

	// Expects a 204 No Content
	err = DecodeNoContent(res)

	if err != nil {
		var srvErr APIError
		if errors.As(err, &srvErr) {
			a.client.logger.Debug("API server returned error response.", "error", srvErr)
			return fmt.Errorf("error from BC API: %w", srvErr)
		}

		a.client.logger.Debug("Failed to decode response.", "error", err)
		return fmt.Errorf("failed to decode response: %w", err)
	}
	a.client.logger.Debug("Succesfully deleted record.", "id", id)

	return nil
}
