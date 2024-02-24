package bc

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Client is used to send ana receive HTTP requests/responses to the
// API server. There should be one client created per publisher/group/version
// combination as these can each have their own schemas. Clients can
// share the authClient if they are using the same scope.
type Client struct {
	authClient TokenGetter
	baseClient *http.Client
	baseURL    *url.URL
	common     bool
	config     ClientConfig
	logger     *slog.Logger
}

// The required configuration options for the Client.
// Meets the Validator interface.
type ClientConfig struct {
	// TenantID is the Entra Tenant ID for the organization.
	TenantID GUID
	// CompanyID is the BC company within the environment.
	CompanyID GUID
	// Environment must be a non-empty string.
	Environment string
	// APIEndpoint must be either "v2.0" or the format
	//"<publisher>/<group>/<version>".
	APIEndpoint string
}

// Validates that the params are all in correct format.
func (p ClientConfig) Validate() error {
	var errs error

	if err := p.TenantID.Validate(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid TenantID: %w", err))
	}

	if err := p.CompanyID.Validate(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid CompanyID: %w", err))
	}

	if err := stringNotEmpty(p.Environment); err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid Environment: %w", err))
	}

	if p.APIEndpoint == "v2.0" {
		return errs
	}

	segmentCount := len(strings.Split(p.APIEndpoint, "/"))
	if segmentCount == 3 {
		return errs
	}

	errs = errors.Join(errs, fmt.Errorf("invalid APIEndpoint: must equal %q or have 3 path segments", "v2.0"))

	return errs
}

// New client takes the mandatory ClientConfig params, a TokenGetter, and
// optional configuration using the functional options pattern.
// Logger and HTTPClient will be set to defaults if not optionally set.
func NewClient(config ClientConfig, authClient TokenGetter, opts ...ClientOptionFunc) (*Client, error) {

	// Validate params
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: \n%w", err)
	}

	// Create the base URL
	baseURL, err := BuildBaseURL(config)
	if err != nil {
		return nil, err
	}

	// Check if it is a common endpoint (v2.0 for now)
	// Common endpoints have different path structure
	isCommon := config.APIEndpoint == "v2.0"

	client := &Client{
		baseURL:    baseURL,
		authClient: authClient,
		config:     config,
		common:     isCommon,
	}

	// Apply the optional functions to the client
	setClientOptions(client, opts)

	return client, nil
}

// APIEndpoint returns the API endpoint, either the version for a common endpoint
// or the <publisher>/<group>/<version> if an extension API.
func (c *Client) APIEndpoint() string {
	return c.config.APIEndpoint
}

// IsCommon returns true if it is a common service endpoint.
func (c *Client) IsCommon() bool {
	return c.common
}

// Config returns ClientConfig for this instance.
func (c *Client) Config() ClientConfig {
	return c.config
}

// BaseClient returns the baseClient *http.Client.
func (c *Client) BaseClient() *http.Client {
	return c.baseClient
}

// BaseClient returns the logger *slog.Logger.
func (c *Client) Logger() *slog.Logger {
	return c.logger
}
