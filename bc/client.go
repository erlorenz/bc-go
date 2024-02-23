package bc

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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

// The manadatory configuration params for the Client.
type ClientConfig struct {
	TenantID    GUID   `validate:"required,uuid"`
	CompanyID   GUID   `validate:"required,uuid"`
	Environment string `validate:"required"`
	APIEndpoint string `validate:"required"`
}

// Validates that the params are all in correct format.
func (p ClientConfig) Validate() error {
	err := validateStruct(p)
	if err != nil {
		// TODO: make the format nicer
		return err
	}
	return err
}

// New client takes the mandatory ClientConfig params, a TokenGetter, and
// optional configuration using the functional options pattern.
// Logger and HTTPClient will be set to defaults if not optionally set.
func NewClient(config ClientConfig, authClient TokenGetter, opts ...ClientOptionFunc) (*Client, error) {

	// Validate params
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: \n%w", err)
	}

	// Check if it is a common endpoint (v2.0 for now)
	// Common endpoints have different path structure
	common := config.APIEndpoint == "v2.0"

	// Create the client
	client := &Client{
		authClient: authClient,
		config:     config,
		common:     common,
	}

	// Range through the options and set them to the client
	setClientOptions(client, opts)

	// Create the base URL
	baseURL, err := BuildBaseURL(config)
	if err != nil {
		return nil, err
	}
	client.baseURL = baseURL

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
