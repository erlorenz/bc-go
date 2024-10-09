package bc

import (
	"cmp"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

var Version = "0.14.0"

// Client is used to send and receive HTTP requests/responses to the
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
	TenantID string
	// CompanyID is the BC company within the environment.
	CompanyID string
	// Environment must be a non-empty string.
	Environment string
	// APIEndpoint must be either "v2.0" or the format
	//"<publisher>/<group>/<version>".
	APIEndpoint string
	// ClientID is also known as the application ID.
	ClientID string
	//ClientSecret is the MSAL client secret for the application.
	ClientSecret string
}

// Validates that the params are all in correct format.
func (cc ClientConfig) Validate() error {
	var errs []string

	if _, err := uuid.Parse(cc.TenantID); err != nil {
		errs = append(errs, fmt.Sprintf("TenantID: %s", err))
	}

	if _, err := uuid.Parse(cc.CompanyID); err != nil {
		errs = append(errs, fmt.Sprintf("CompanyID: %s", err))
	}

	if err := stringNotEmpty(cc.Environment); err != nil {
		errs = append(errs, fmt.Sprintf("Environment: %s", err))
	}

	if _, err := uuid.Parse(cc.ClientID); err != nil {
		errs = append(errs, fmt.Sprintf("ClientID: %s", err))
	}

	if err := stringNotEmpty(cc.ClientSecret); err != nil {
		errs = append(errs, fmt.Sprintf("ClientSecret: %s", err))
	}

	if cc.APIEndpoint == "v2.0" {
		if len(errs) > 0 {
			return fmt.Errorf("validate config: [%s]", strings.Join(errs, ", "))
		}
		return nil
	}

	segmentCount := len(strings.Split(cc.APIEndpoint, "/"))
	if segmentCount == 3 {
		if len(errs) > 0 {
			return fmt.Errorf("validate config: [%s]", strings.Join(errs, ", "))
		}
		return nil
	}

	errs = append(errs, fmt.Sprintf("APIEndpoint: must equal %q or have 3 path segments", "v2.0"))

	return fmt.Errorf("validate config: [%s]", strings.Join(errs, ", "))
}

// NewClient creates a [Client] with configuration params and optional configuration with functional options.
// Available options are [WithAuthClient], [WithLogger], [WithHTTPClient].
func NewClient(config ClientConfig, opts ...ClientOption) (*Client, error) {

	// Validate params
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: \n%w", err)
	}

	// Create the base URL
	baseURL, err := BuildBaseURL(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		baseURL: baseURL,
		config:  config,
	}

	// Apply the optional functions to the client
	for _, opt := range opts {
		opt(client)
	}

	if client.authClient == nil {
		ac, err := NewAuth(config.TenantID, config.ClientID, config.ClientSecret)
		if err != nil {
			return nil, err
		}
		client.authClient = ac
	}

	client.logger = cmp.Or(client.logger, slog.Default())
	client.baseClient = cmp.Or(client.baseClient, &http.Client{Timeout: 20 * time.Second})

	return client, nil
}

// APIEndpoint returns the API endpoint, either the version for a common endpoint
// or the <publisher>/<group>/<version> if an extension API.
func (c *Client) APIEndpoint() string {
	return c.config.APIEndpoint
}

// IsCommon returns true if it is a common service endpoint.
func (c *Client) IsCommon() bool {
	return c.config.APIEndpoint == "v2.0"
}

// Config returns ClientConfig for this instance.
func (c *Client) Config() ClientConfig {
	return c.config
}

// BaseClient returns the baseClient [http.Client].
func (c *Client) BaseClient() *http.Client {
	return c.baseClient
}

// BaseClient returns the logger *slog.Logger.
func (c *Client) Logger() *slog.Logger {
	return c.logger
}
