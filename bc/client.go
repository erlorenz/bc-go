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
	TenantID    GUID
	CompanyID   GUID
	Environment string
	APIEndpoint string
}

// Validates that the params are all in correct format.
func (p ClientConfig) Validate() error {
	err := ValidatorMap{
		"TenantID":    p.TenantID,
		"CompanyID":   p.CompanyID,
		"Environment": NotEmptyString(p.Environment),
		"APIEndpoint": NotEmptyString(p.APIEndpoint),
	}.Validate()

	if err != nil {
		return err
	}
	return nil
}

// New client takes the mandatory ClientConfig params, a TokenGetter, and
// optional configuration using the functional options pattern.
// Logger and HTTPClient will be set to defaults if not optionally set.
func NewClient(config ClientConfig, authClient TokenGetter, opts ...ClientOptionFunc) (*Client, error) {

	// Validate params
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating ClientParams: %w", err)
	}

	// Create ClientOptions and set them from opts
	options := newClientOptions(opts)

	// Check if it is a common endpoint (v2.0 for now)
	// Common endpoints have different path structure
	common := config.APIEndpoint == "v2.0"

	// Create the client
	client := &Client{
		authClient: authClient,
		config:     config,
		logger:     options.logger,
		baseClient: options.httpClient,
		common:     common,
	}

	// Create the base URL
	baseURL, err := buildBaseURL(config)
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
