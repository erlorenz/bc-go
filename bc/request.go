package bc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// HTTP and OData constants
const (
	ContentTypeJSON    = "application/json"
	DataAccessReadOnly = "ReadOnly"
)

// ODataQuery represents OData query parameters.
// Use the ToValues() method to convert to url.Values for the request.
// Zero values are omitted from the query string.
type ODataQuery struct {
	Filter  string   // $filter - filter collection by expression
	Select  []string // $select - select specific fields
	Expand  []string // $expand - expand related entities
	OrderBy string   // $orderby - sort results
	Top     int      // $top - limit number of results
	Skip    int      // $skip - skip N results
	Count   bool     // $count - include total count
}

// ToValues converts the ODataQuery to url.Values for use in HTTP requests.
// Empty/zero values are omitted.
func (q *ODataQuery) ToValues() url.Values {
	values := url.Values{}

	if q.Filter != "" {
		values.Set("$filter", q.Filter)
	}
	if len(q.Select) > 0 {
		values.Set("$select", strings.Join(q.Select, ","))
	}
	if len(q.Expand) > 0 {
		values.Set("$expand", strings.Join(q.Expand, ","))
	}
	if q.OrderBy != "" {
		values.Set("$orderby", q.OrderBy)
	}
	if q.Top > 0 {
		values.Set("$top", strconv.Itoa(q.Top))
	}
	if q.Skip > 0 {
		values.Set("$skip", strconv.Itoa(q.Skip))
	}
	if q.Count {
		values.Set("$count", "true")
	}

	return values
}

// Response wraps an HTTP response with the raw body and BC-specific metadata.
// The raw body is always read and stored during DoRequest, so the response
// can be decoded multiple ways or inspected.
type Response struct {
	// HTTPResponse is the underlying HTTP response
	HTTPResponse *http.Response

	// RawBody contains the raw response body.
	// The body is always read and closed during DoRequest.
	RawBody json.RawMessage

	// BC-specific metadata from response headers
	RequestID string // request-id header
	SessionID string // x-ms-session-id header
}

// DecodeJSON decodes the raw response body as JSON into v.
// The body has already been read and closed during DoRequest.
func (r *Response) DecodeJSON(v any) error {
	if len(r.RawBody) == 0 {
		return nil // Empty body, nothing to decode
	}
	return json.Unmarshal(r.RawBody, v)
}

// RequestOption modifies request behavior (headers, etc.).
type RequestOption func(*http.Request)

// DoRequest executes an HTTP request to Business Central.
// The response body is always read and stored in Response.RawBody, even on success.
//
// Required parameters:
//   - ctx: context for cancellation/timeout
//   - method: HTTP method (GET, POST, PATCH, DELETE)
//   - path: URL path segment appended to baseURL, OR full URL for @odata.nextLink
//     Examples: "customers", "customers(id)", "https://api.../customers?$skiptoken=..."
//   - query: OData query parameters, can be nil (ignored if path is full URL)
//   - body: Request body (nil for GET/DELETE)
//
// Optional parameters via RequestOption:
//   - WithReadOnly(): Use read replica (for list/get where eventual consistency is OK)
//   - WithMaxPageSize(n): Set odata.maxpagesize preference
//   - WithPreferRepresentation(): Return full entity in response (use with POST)
//   - WithPreferMinimal(): Return minimal response
//   - WithETag(etag): Use specific ETag for If-Match
//   - WithHeader(key, value): Set custom header
func (c *Client) DoRequest(
	ctx context.Context,
	method string,
	path string,
	query *ODataQuery,
	body any,
	opts ...RequestOption,
) (*Response, error) {
	// Build full URL
	// Special case: if path is a full URL (e.g., @odata.nextLink), use it as-is
	var fullURL string
	var isOpaqueURL bool

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// This is @odata.nextLink - use as opaque URL per OData spec
		fullURL = path
		isOpaqueURL = true
	} else {
		// Normal case - build URL from baseURL + path
		var err error
		fullURL, err = url.JoinPath(c.baseURL, path)
		if err != nil {
			return nil, fmt.Errorf("building URL: %w", err)
		}
	}

	// Marshal body if present
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add query parameters (skip if using opaque URL - it already has query params)
	if query != nil && !isOpaqueURL {
		req.URL.RawQuery = query.ToValues().Encode()
	}

	// Method-specific headers
	// Note: Accept, Authorization, and User-Agent are set by bcTransport

	// Set Content-Type only if we have a body
	if body != nil {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}

	// Set If-Match for destructive operations (can be overridden by WithETag)
	if method == http.MethodPatch || method == http.MethodPut || method == http.MethodDelete {
		req.Header.Set("If-Match", "*")
	}

	// Apply functional options (can override defaults)
	for _, opt := range opts {
		opt(req)
	}

	// Execute request (transport adds Authorization and User-Agent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	// Always read and close the body
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, c.parseError(resp.StatusCode, resp.Status, rawBody)
	}

	// Wrap response with BC metadata and raw body
	return &Response{
		HTTPResponse: resp,
		RawBody:      json.RawMessage(rawBody),
		RequestID:    resp.Header.Get("request-id"),
	}, nil
}

// parseError extracts error details from OData error responses.
func (c *Client) parseError(statusCode int, status string, body []byte) error {
	// TODO: Parse OData error format properly
	// OData errors look like:
	// {
	//   "error": {
	//     "code": "BadRequest",
	//     "message": "The field 'email' is required"
	//   }
	// }

	// For now, return basic error with body
	if len(body) > 0 {
		return fmt.Errorf("HTTP %d: %s - %s", statusCode, status, string(body))
	}
	return fmt.Errorf("HTTP %d: %s", statusCode, status)
}

// WithMaxPageSize sets the odata.maxpagesize preference for server-driven paging.
func WithMaxPageSize(size int) RequestOption {
	return func(req *http.Request) {
		addPreference(req, fmt.Sprintf("odata.maxpagesize=%d", size))
	}
}

// WithETag sets a specific ETag for optimistic concurrency control.
// Overrides the default If-Match: * behavior.
func WithETag(etag string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("If-Match", etag)
	}
}

// WithHeader sets a custom header on the request.
func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

// WithAddHeader adds a header value (doesn't overwrite existing).
func WithAddHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Add(key, value)
	}
}

// WithDataAccessReadOnly sets Data-Access-Intent to ReadOnly, directing reads to a replica.
// Use this for list and get operations where eventual consistency is acceptable.
// Do NOT use immediately after an update, as the replica may not have the changes yet.
func WithDataAccessReadOnly() RequestOption {
	return func(req *http.Request) {
		req.Header.Set("Data-Access-Intent", DataAccessReadOnly)
	}
}

// addPreference is a helper that properly appends to the Prefer header.
func addPreference(req *http.Request, pref string) {
	existing := req.Header.Get("Prefer")
	if existing != "" {
		pref = existing + ", " + pref
	}
	req.Header.Set("Prefer", pref)
}
