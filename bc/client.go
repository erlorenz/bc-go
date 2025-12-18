package bc

import (
	"cmp"
	"fmt"
	"log/slog"
	"net/http"
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
