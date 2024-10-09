package bc

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

// dksdlfjdsja
// BuildBaseURL builds the BaseURL from the ClientConfig.
// It uses the structure
// "https://api.businesscentral.dynamics.com/v2.0/{tenantID}/{environment}/api/{APIendpoint}/companies({companyID})"
func BuildBaseURL(cfg ClientConfig) (*url.URL, error) {

	// All BC APIs use this same prefix.
	staticPrefix := "https://api.businesscentral.dynamics.com/v2.0"

	// Specific to this Client
	dynamicURL := fmt.Sprintf("/%s/%s/api/%s/companies(%s)", cfg.TenantID, cfg.Environment, cfg.APIEndpoint, cfg.CompanyID)

	// Final combination
	baseURLstring := staticPrefix + dynamicURL
	baseURL, err := url.Parse(baseURLstring)
	if err != nil {
		return &url.URL{}, fmt.Errorf("error building BaseURL: %w", err)
	}

	return baseURL, nil
}

// BuildRequestURL builds a URL to be used in an http.Request.
// It uses the structure
// https://api.businesscentral.dynamics.com/v2.0/{tenantID}/{environment}/api/{APIendpoint}/companies({companyID})/{entitySet}({recordID})?{queryParams}
func BuildRequestURL(baseURL url.URL, entitySet string, recordID uuid.UUID, queryParams QueryParams) url.URL {

	newURL := baseURL
	// Don't forget the slash in between
	entitySetPath := "/" + entitySet
	newURL.Path += entitySetPath

	// Add recordID if exists
	if recordID != uuid.Nil {
		newURL.Path += fmt.Sprintf("(%s)", recordID)
	}

	// Build query params, skip if empty
	query := url.Values{}
	for k, v := range queryParams {
		if v == "" {
			continue
		}
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
// const pathIndexEntitySetName = 9

// const pathIndexCommonVersion = 3
// const pathIndexCommonCompaniesSegment = 6
// const pathIndexCommonEntitySetName = 7
