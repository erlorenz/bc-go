package bc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type APIPage[T Validator] struct {
	entitySetName string
	baseFilter    string
	client        Client
	expand        string
}

func (a *APIPage[T]) SetExpand(expand string) {
	a.expand = expand
}

func (a *APIPage[T]) SetBaseFilter(filter string) {
	a.baseFilter = filter
}

func (a *APIPage[T]) Get(ctx context.Context, id GUID) (T, error) {
	var v T
	opts := MakeRequestOptions{
		Method:        http.MethodGet,
		EntitySetName: a.entitySetName,
		RecordID:      id,
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
	if err != nil {
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
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
