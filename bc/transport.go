package bc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// bcTransport is an http.RoundTripper that adds Business Central-specific behavior.
// It wraps a base transport and adds:
//   - Authentication (Bearer token via azcore.TokenCredential)
//   - User-Agent header
//
// The base transport is what the user provides (which may be OTelHTTP or http.DefaultTransport).
// Request flow: bcTransport → base (user's OTelHTTP) → network
//
// Future: retry logic will wrap this transport (retry → bcTransport → otelhttp → network)
type bcTransport struct {
	base      http.RoundTripper
	cred      azcore.TokenCredential
	userAgent string
}

// newBCTransport creates a new Business Central transport that wraps the base transport.
func newBCTransport(base http.RoundTripper, cred azcore.TokenCredential, userAgent string) *bcTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &bcTransport{
		base:      base,
		cred:      cred,
		userAgent: userAgent,
	}
}

// RoundTrip implements http.RoundTripper.
// It adds authentication and user agent headers, then delegates to the base transport.
// The request is cloned to avoid modifying the original, which is important for retry logic.
func (t *bcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	// This is important for retry logic which may reuse the same request
	req = req.Clone(req.Context())

	// Get bearer token
	token, err := t.getBearerToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("getting bearer token: %w", err)
	}

	// Set BC-specific headers
	req.Header.Set("Authorization", token)
	req.Header.Set("User-Agent", t.userAgent)

	// Delegate to base transport (user's OTelHTTP or DefaultTransport)
	return t.base.RoundTrip(req)
}

// getBearerToken gets the access token from the credential and formats it as a Bearer token.
func (t *bcTransport) getBearerToken(ctx context.Context) (string, error) {
	token, err := t.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://api.businesscentral.dynamics.com/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("getting access token: %w", err)
	}

	return fmt.Sprintf("Bearer %s", token.Token), nil
}
