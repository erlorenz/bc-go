package bcgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const ContentTypeJSON = "application/json"
const NoODATAMetadata = "odata.metadata=none"
const DataAccessReadOnly = "ReadOnly"

// This is so that it doesn't return a bunch of OData stuff. It's semicolon separated.
var AcceptJSONNoMetadata = strings.Join([]string{ContentTypeJSON, NoODATAMetadata}, ";")

// MakeRequestOptions are the unique options for the http.Request.
type MakeRequestOptions struct {
	method        string
	entitySetName string
	queryParams   map[string]string
	body          any
}

// QueryParams are used to build the http.Request url.
type QueryParams map[string]string

// MakeRequest is the base method that creates and executes the HTTP request.
// It has the same return as http.RequestWithContext.
func (c *Client) NewRequest(ctx context.Context, opts MakeRequestOptions) (*http.Request, error) {

	// Build the full URL string
	urlString := buildRequestURL(*c.baseURL, opts.entitySetName, opts.queryParams)

	// Marshall JSON
	body, err := json.Marshal(opts.body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal body %s: %w", opts.body, err)
	}

	// Create Request
	req, err := http.NewRequestWithContext(ctx, opts.method, string(urlString), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating NewRequestWithContext: %w", err)
	}

	// Add the Authorization header for each request
	authHeader, err := getAuthHeader(ctx, c.authClient)
	if err != nil {
		return nil, fmt.Errorf("error creating Auth header: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	// Add this header so it doesn't return all the odata junk
	req.Header.Set("Accept", AcceptJSONNoMetadata)

	// Use ReadOnly for GET
	if opts.method == http.MethodGet {
		req.Header.Set("Data-Access-Intent", DataAccessReadOnly)
	}

	// Use JSON for POST, PUT, PATCH
	if opts.method == http.MethodPost || opts.method == http.MethodPut || opts.method == http.MethodPatch {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}

	// Use If-Match for requests that modify existing
	if opts.method == http.MethodDelete || opts.method == http.MethodPut || opts.method == http.MethodPatch {
		req.Header.Set("If-Match", "*")
	}

	return req, nil

}

// getAuthHeader gets the AccessToken and creates a Bearer token.
func getAuthHeader(ctx context.Context, tg TokenGetter) (string, error) {
	accessToken, err := tg.GetToken(ctx)
	if err != nil {
		return "", fmt.Errorf("error adding auth header: %w", err)
	}

	return fmt.Sprintf("Bearer %s", accessToken), nil

}
