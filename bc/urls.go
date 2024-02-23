package bc

import (
	"fmt"
	"net/url"
)

// BuildBaseURL builds the BaseURL from the ClientConfig.
func BuildBaseURL(cfg ClientConfig) (*url.URL, error) {
	// All BC APIs use this
	staticURL := "https://api.businesscentral.dynamics.com/v2.0"

	// Specific to this Client
	dynamicURL := fmt.Sprintf("/%s/%s/api/%s/companies(%s)", cfg.TenantID, cfg.Environment, cfg.APIEndpoint, cfg.CompanyID)

	// Final combination
	rawBaseURL := staticURL + dynamicURL
	baseURL, err := url.ParseRequestURI(rawBaseURL)
	if err != nil {
		return &url.URL{}, fmt.Errorf("error building BaseURL: %w", err)
	}

	return baseURL, nil
}

func BuildRequestURL(baseURL url.URL, entitySet string, recordID GUID, queryParams QueryParams) url.URL {

	newURL := baseURL
	// Don't forget the slash in between
	entitySetPath := "/" + entitySet
	newURL.Path += entitySetPath

	// Add recordID if exists
	if recordID != "" {
		newURL.Path += fmt.Sprintf("(%s)", recordID)
	}

	// Build query params
	query := url.Values{}
	for k, v := range queryParams {
		query.Set(k, v)
	}
	newURL.RawQuery = query.Encode()

	return newURL
}
