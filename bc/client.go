package bc

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/google/uuid"
)

// Version is the semantic version of the bc-go library.
// Update this constant when releasing new versions.
const Version = "0.16.0"

// Client is used to send and receive HTTP requests/responses to the
// API server. There should be one client created per Route (publisher/group/version)
// and CompanyID combination as these can each have their own schemas. Clients can
// share the credential if they are using the same scope.
type Client struct {
	cred       azcore.TokenCredential
	httpClient *http.Client
	baseURL    string
	route      string
	logger     *slog.Logger
	userAgent  string
}

// ClientOptions contains optional configuration for the Client.
// All fields are optional and have sensible defaults.
type ClientOptions struct {
	// RootURL overrides the default root URL (primarily for testing).
	// Default: "https://api.businesscentral.dynamics.com"
	// If provided, should be just the scheme + host (e.g., "http://localhost:8080")
	RootURL string

	// HTTPClient allows providing a custom HTTP client.
	// If nil, a default client with 20s timeout is used.
	HTTPClient *http.Client

	// Logger for structured logging.
	// If nil, slog.Default() is used.
	Logger *slog.Logger

	// UserAgentSuffix is appended to the default user agent string.
	// Format: "bc-go/<version> <suffix>"
	UserAgentSuffix string
}

// NewClient creates a new Business Central API client.
//
// Required parameters:
//   - cred: Azure credential for authentication (e.g., from azidentity package)
//   - tenantID: Entra tenant ID (GUID format)
//   - environment: BC environment name (e.g., "Production", "Sandbox")
//   - companyID: BC company ID (GUID format)
//   - route: API route, either "v2.0" for common endpoints or "publisher/group/version" for extensions
//
// Optional parameters via options:
//   - RootURL: Override default API host (for testing)
//   - HTTPClient: Custom HTTP client
//   - Logger: Custom structured logger
//   - UserAgentSuffix: Additional user agent identification
func NewClient(
	cred azcore.TokenCredential,
	tenantID string,
	environment string,
	companyID string,
	route string,
	options *ClientOptions,
) (*Client, error) {
	// Initialize options if nil
	if options == nil {
		options = &ClientOptions{}
	}

	// Validate required parameters
	var errs []string

	if cred == nil {
		errs = append(errs, "cred cannot be nil")
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		errs = append(errs, fmt.Sprintf("tenantID: %s", err))
	}

	if _, err := uuid.Parse(companyID); err != nil {
		errs = append(errs, fmt.Sprintf("companyID: %s", err))
	}

	if environment == "" {
		errs = append(errs, "environment cannot be empty")
	}

	// Validate route format
	if route != "v2.0" {
		segmentCount := len(strings.Split(route, "/"))
		if segmentCount != 3 {
			errs = append(errs, fmt.Sprintf("route: must be %q or have 3 path segments (publisher/group/version)", "v2.0"))
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid parameters: [%s]", strings.Join(errs, ", "))
	}

	// Apply defaults with cmp.Or
	rootURL := cmp.Or(options.RootURL, "https://api.businesscentral.dynamics.com")
	logger := cmp.Or(options.Logger, slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError + 1})))

	// Build base URL
	baseURL := fmt.Sprintf("%s/v2.0/%s/%s/api/%s/companies(%s)",
		rootURL, tenantID, environment, route, companyID)

	// Build user agent
	userAgent := "bc-go/" + Version
	if options.UserAgentSuffix != "" {
		userAgent += " " + options.UserAgentSuffix
	}

	// Get base transport from user's client or use default
	var baseTransport http.RoundTripper = http.DefaultTransport
	if options.HTTPClient != nil && options.HTTPClient.Transport != nil {
		baseTransport = options.HTTPClient.Transport
	}

	// Wrap base transport with BC transport (auth + user-agent)
	bcTransport := newBCTransport(baseTransport, cred, userAgent)

	// Create HTTP client with wrapped transport
	// Preserve user's timeout, jar, and redirect policy if provided
	httpClient := &http.Client{
		Transport: bcTransport,
		Timeout:   20 * time.Second,
	}
	if options.HTTPClient != nil {
		httpClient.Timeout = options.HTTPClient.Timeout
		httpClient.CheckRedirect = options.HTTPClient.CheckRedirect
		httpClient.Jar = options.HTTPClient.Jar
	}

	return &Client{
		cred:       cred,
		httpClient: httpClient,
		baseURL:    baseURL,
		route:      route,
		logger:     logger,
		userAgent:  userAgent,
	}, nil
}

// NewRequest creates a new bc.Request from components (like http.NewRequest).
// This is the low-level API for building requests manually.
//
// Parameters:
//   - method: HTTP method (GET, POST, PATCH, DELETE)
//   - path: URL path segment to append to client's baseURL, OR full URL for @odata.nextLink
//   - query: OData query parameters (can be nil, ignored if path is full URL)
//   - body: Request body to JSON marshal (can be nil)
//   - opts: Optional RequestOptions for setting headers/preferences
//
// Returns a Request with a fully-built URL ready for execution.
func NewRequest(c *Client, method, path string, query *ODataQuery, body any, opts ...RequestOption) (*Request, error) {
	var fullURL string

	// Check if path is a full URL (opaque @odata.nextLink)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Opaque URL - use as-is
		fullURL = path
	} else {
		// Normal path - parse baseURL and join
		baseURL, err := url.Parse(c.baseURL)
		if err != nil {
			return nil, fmt.Errorf("parsing base URL: %w", err)
		}

		joined := baseURL.JoinPath(path)

		// Add query parameters if provided
		if query != nil {
			joined.RawQuery = query.ToValues().Encode()
		}

		fullURL = joined.String()
	}

	// Marshal body if present
	var bodyBytes json.RawMessage
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyBytes = b
	}

	// Create request with headers
	req := &Request{
		Method: method,
		URL:    fullURL,
		Body:   bodyBytes,
		Header: make(http.Header),
	}

	// Set default headers
	if body != nil {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}

	if method == http.MethodPatch || method == http.MethodPut || method == http.MethodDelete {
		req.Header.Set("If-Match", "*")
	}

	// Apply RequestOptions
	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}

// Do executes a bc.Request and returns a Response (like http.Client.Do).
// This is the primary API for executing single requests.
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// Create body reader
	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	// Use headers from bc.Request
	httpReq.Header = req.Header

	// Execute request (transport adds Authorization, User-Agent, Accept)
	resp, err := c.httpClient.Do(httpReq)
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

	// Wrap response with BC metadata
	return &Response{
		HTTPResponse: resp,
		RawBody:      json.RawMessage(rawBody),
		RequestID:    resp.Header.Get("request-id"),
	}, nil
}

// DoRequest is a convenience method that combines NewRequest + Do.
// It exists for backward compatibility and as a simple API for one-off requests.
func (c *Client) DoRequest(
	ctx context.Context,
	method, path string,
	query *ODataQuery,
	body any,
	opts ...RequestOption,
) (*Response, error) {
	// Build request
	req, err := NewRequest(c, method, path, query, body, opts...)
	if err != nil {
		return nil, err
	}

	// Execute request
	return c.Do(ctx, req)
}

// Unexported methods for testing

func (c *Client) getBaseURL() string {
	return c.baseURL
}

func (c *Client) getRoute() string {
	return c.route
}

func (c *Client) getUserAgent() string {
	return c.userAgent
}

func (c *Client) getHTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) getLogger() *slog.Logger {
	return c.logger
}
