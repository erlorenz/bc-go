package mock

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// Credential is a fake azcore.TokenCredential for testing.
// It returns a static fake token without making any network calls.
type Credential struct {
	// Token is the fake token to return. Defaults to "FAKE_TOKEN" if empty.
	Token string
	// Err is the error to return. If nil, no error is returned.
	Err error
}

// GetToken implements azcore.TokenCredential.
// It returns a fake access token for testing purposes.
func (fc *Credential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if fc.Err != nil {
		return azcore.AccessToken{}, fc.Err
	}

	token := fc.Token
	if token == "" {
		token = "FAKE_TOKEN"
	}

	return azcore.AccessToken{
		Token:     token,
		ExpiresOn: time.Now().Add(1 * time.Hour),
	}, nil
}
