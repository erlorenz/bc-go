package bcgo

import (
	"cmp"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Client struct {
	authClient TokenGetter
	baseClient *http.Client
	baseURL    URLString
	config     ClientConfig
	logger     *slog.Logger
}

type ClientConfig struct {
	TenantID     GUID
	CompanyID    GUID
	Environment  string
	APIPublisher string
	APIGroup     string
	APIVersion   string
}

func (p ClientConfig) Validate() error {
	err := ValidatorMap{
		"TenantID":     p.TenantID,
		"CompanyID":    p.CompanyID,
		"Environment":  NotEmptyString(p.Environment),
		"APIPublisher": NotEmptyString(p.APIPublisher),
		"APIGroup":     NotEmptyString(p.APIGroup),
		"APIVersion":   NotEmptyString(p.APIVersion),
	}.Validate()

	if err != nil {
		return err
	}
	return nil
}

// New takes
func NewClient(config ClientConfig, authClient TokenGetter, opts ...func(*Client)) (*Client, error) {

	// Validate params
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating ClientParams: %w", err)
	}

	// Create the client
	client := &Client{
		authClient: authClient,
		config:     config,
		baseClient: nil,
		logger:     nil,
	}

	// Create the base URL
	baseURL, err := buildBaseURL(config)
	if err != nil {
		return nil, err
	}
	client.baseURL = baseURL

	// Apply options
	for _, o := range opts {
		o(client)
	}

	// Create a default http client and logger if not provided
	client.baseClient = cmp.Or(client.baseClient, &http.Client{
		Timeout: time.Second * 15,
	})

	client.logger = cmp.Or(client.logger, slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	return client, nil
}

// PubGroup returns the API Publisher and API Group
// in the format <publisher>/<group>
func (c *Client) PubGroup() string {
	return fmt.Sprintf("%s/%s", c.config.APIPublisher, c.config.APIGroup)
}

// Config returns ClientConfig for this instance
func (c *Client) Config() ClientConfig {
	return c.config
}

// buildBaseURL builds the BaseURL from the ClientConfig.
func buildBaseURL(cfg ClientConfig) (URLString, error) {
	// All BC APIs use this
	staticURL := "https://api.businesscentral.dynamics.com/v2.0"

	// Specific to this Client
	apiPathSegments := fmt.Sprintf("%s/%s/%s", cfg.APIPublisher, cfg.APIGroup, cfg.APIVersion)
	dynamicURL := fmt.Sprintf("/%s/%s/api/%s/companies(%s)", cfg.TenantID, cfg.Environment, apiPathSegments, cfg.CompanyID)

	// Final combination
	baseURL := URLString(staticURL + dynamicURL)
	if err := baseURL.Validate(); err != nil {
		return "", fmt.Errorf("error building BaseURL: %w", err)
	}

	return baseURL, nil
}
