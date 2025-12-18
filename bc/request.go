package bc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

const ContentTypeJSON = "application/json"
const NoODATAMetadata = "odata.metadata=none"
const DataAccessReadOnly = "ReadOnly"

// This is the "Accept" header value to return JSON without the OData metadata.
// It's semicolon separated. Included in all requests.
var AcceptJSONNoMetadata = strings.Join([]string{ContentTypeJSON, NoODATAMetadata}, ";")

// MakeRequestOptions are the unique options for the http.Request.
type RequestOptions struct {
	Method        string
	EntitySetName string
	RecordID      uuid.UUID
	QueryParams   QueryParams
	Body          any
}

// Validate checks all the fields for invalid combinations or values.
func (r RequestOptions) Validate() error {
	var errs []string

	// Validate method and entity set to be required.
	if r.Method == "" {
		errs = append(errs, fmt.Sprintf("invalid method: %s", r.Method))
	}

	if r.EntitySetName == "" {
		errs = append(errs, fmt.Sprintf("invalid entitysetname: %s", r.Method))
	}

	// If body exist the method cant be get or delete
	if r.Body != nil {
		if r.Method == http.MethodGet || r.Method == http.MethodDelete {
			errs = append(errs, "invalid combination: cannot have body with GET or DELETE method")
		}
	}
	// Cannot have filter query params with anything but GET
	if r.QueryParams != nil && r.QueryParams["$filter"] != "" {
		if r.Method != http.MethodGet {
			errs = append(errs, fmt.Sprintf("invalid combination: cannot have $filter query param with method %s", r.Method))
		}
	}
	if r.Method == http.MethodPatch && r.RecordID == uuid.Nil {
		errs = append(errs, "invalid combination: cannot have method PATCH with no RecordID")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid requestoptions: [ %s ]", strings.Join(errs, ", "))
	}
	return nil
}

// QueryParams are used to build the http.Request url.
type QueryParams map[string]string

// NewRequest is the base method that creates the http.Request.
// It has the same return as http.RequestWithContext.
func (c *Client) NewRequest(ctx context.Context, opts RequestOptions) (*http.Request, error) {

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Build the URL path with entity set and optional record ID
	pathSegment := opts.EntitySetName
	if opts.RecordID != uuid.Nil {
		pathSegment += fmt.Sprintf("(%s)", opts.RecordID)
	}

	// Join base URL with path segment
	fullURL, err := url.JoinPath(c.baseURL, pathSegment)
	if err != nil {
		return nil, fmt.Errorf("building URL: %w", err)
	}

	// Marshall JSON
	var body io.Reader
	if opts.Body != nil {
		b, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal body %s: %w", opts.Body, err)
		}
		body = bytes.NewReader(b)
	}

	// Create Request
	req, err := http.NewRequestWithContext(ctx, opts.Method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating new request: %w", err)
	}

	// Add query params
	if opts.QueryParams != nil {
		q := req.URL.Query()
		for k, v := range opts.QueryParams {
			if v != "" {
				q.Set(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	// Add BC-specific headers
	// Note: Authorization and User-Agent are added by bcTransport
	req.Header.Set("Accept", AcceptJSONNoMetadata)

	// Use ReadOnly for GET
	if opts.Method == http.MethodGet {
		req.Header.Set("Data-Access-Intent", DataAccessReadOnly)
	}

	// Use JSON for POST, PUT, PATCH
	if opts.Method == http.MethodPost || opts.Method == http.MethodPut || opts.Method == http.MethodPatch {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}

	// Use If-Match for POST, PUT, PATCH, DELETE
	if opts.Method == http.MethodDelete || opts.Method == http.MethodPut || opts.Method == http.MethodPatch {
		req.Header.Set("If-Match", "*")
	}

	return req, nil
}

// Do calls Do on the HTTP client.
func (c *Client) Do(r *http.Request) (*http.Response, error) {
	return c.httpClient.Do(r)
}
