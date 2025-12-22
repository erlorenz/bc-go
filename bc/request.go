package bc

import (
	"encoding/json"
	"fmt"
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
	if q.Count {
		values.Set("$count", "true")
	}

	return values
}

// Request represents a deferred HTTP request that can be executed immediately
// or serialized for batch operations. The URL is fully built and ready for use.
type Request struct {
	// Method is the HTTP method (GET, POST, PATCH, DELETE)
	Method string

	// URL is the fully-built request URL (absolute or relative for batch)
	URL string

	// Body is the JSON-serialized request body (nil for GET/DELETE)
	Body json.RawMessage

	// Header contains request headers
	Header http.Header
}

// Clone creates a deep copy of the Request.
// Useful for batch operations where the same request template is used multiple times.
func (r *Request) Clone() *Request {
	clone := &Request{
		Method: r.Method,
		URL:    r.URL,
	}

	// Deep copy body
	if r.Body != nil {
		clone.Body = make(json.RawMessage, len(r.Body))
		copy(clone.Body, r.Body)
	}

	// Deep copy headers
	if r.Header != nil {
		clone.Header = make(http.Header, len(r.Header))
		for k, v := range r.Header {
			clone.Header[k] = append([]string(nil), v...)
		}
	}

	return clone
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

// RequestOption modifies a bc.Request (headers, etc.).
type RequestOption func(*Request)

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
	return func(req *Request) {
		addPreference(req, fmt.Sprintf("odata.maxpagesize=%d", size))
	}
}

// WithETag sets a specific ETag for optimistic concurrency control.
// Overrides the default If-Match: * behavior.
func WithETag(etag string) RequestOption {
	return func(req *Request) {
		req.Header.Set("If-Match", etag)
	}
}

// WithHeader sets a custom header on the request.
func WithHeader(key, value string) RequestOption {
	return func(req *Request) {
		req.Header.Set(key, value)
	}
}

// WithAddHeader adds a header value (doesn't overwrite existing).
func WithAddHeader(key, value string) RequestOption {
	return func(req *Request) {
		req.Header.Add(key, value)
	}
}

// WithDataAccessReadOnly sets Data-Access-Intent to ReadOnly, directing reads to a replica.
// Use this for list and get operations where eventual consistency is acceptable.
// Do NOT use immediately after an update, as the replica may not have the changes yet.
func WithDataAccessReadOnly() RequestOption {
	return func(req *Request) {
		req.Header.Set("Data-Access-Intent", DataAccessReadOnly)
	}
}

// addPreference is a helper that properly appends to the Prefer header.
func addPreference(req *Request, pref string) {
	existing := req.Header.Get("Prefer")
	if existing != "" {
		pref = existing + ", " + pref
	}
	req.Header.Set("Prefer", pref)
}
