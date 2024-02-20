package bc

import (
	"fmt"
	"net/url"
)

// buildBaseURL builds the BaseURL from the ClientConfig.
func buildBaseURL(cfg ClientConfig) (*url.URL, error) {
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

func buildRequestURL(baseURL url.URL, entitySet string, recordID GUID, queryParams QueryParams) url.URL {
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

// const pathIndexTenant = 2
// const pathIndexEnvironment = 3
// const pathIndexPublisher = 5
// const pathIndexGroup = 6
// const pathIndexVersion = 7
// const pathIndexCompaniesSegment = 8
const pathIndexEntitySetName = 9

// const pathIndexCommonVersion = 3
// const pathIndexCommonCompaniesSegment = 6
const pathIndexCommonEntitySetName = 7
