package bc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// AuthClient is used to retrieve an AccessToken.
// Implements the TokenGetter interface.
type AuthClient struct {
	client confidential.Client
	scopes []string
	logger *slog.Logger
}

// AuthParams are used to create an AuthClient. Implements the Validator interface.
type AuthParams struct {
	TenantID     GUID
	ClientID     GUID
	ClientSecret string
	Logger       *slog.Logger
}

func (p AuthParams) Validate() error {

	err := ValidatorMap{
		"TenantID":     p.TenantID,
		"ClientID":     p.ClientID,
		"ClientSecret": NotEmptyString(p.ClientSecret),
	}.Validate()

	if err != nil {
		return err
	}
	return nil
}

// AccessToken is used in the Authorization header of requests.
type AccessToken string

// TokenGetter represents a client that retrieves
// an AccessToken.
type TokenGetter interface {
	GetToken(context.Context) (AccessToken, error)
}

// NewAuthClient validates the AuthParams and creates a new AuthClient.
func NewAuthClient(config AuthParams) (*AuthClient, error) {

	// Validate config values
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating AuthParams: [%w]", err)
	}

	cred, err := confidential.NewCredFromSecret(config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("error creating MSAL confidential.Credential: %w", err)
	}

	authority := "https://login.microsoft.com/" + string(config.TenantID)

	confidentialClient, err := confidential.New(authority, string(config.ClientID), cred)
	if err != nil {
		return nil, fmt.Errorf("error creating MSAL confidential.Client: %w", err)
	}

	// Don't think there is any reason to use a different one.
	// Can have this as a config param if ever need to.
	scopes := []string{"https://api.businesscentral.dynamics.com/.default"}

	return &AuthClient{
		client: confidentialClient,
		scopes: scopes,
		logger: config.Logger,
	}, nil

}

func (ac *AuthClient) GetToken(ctx context.Context) (AccessToken, error) {

	ac.logger.Debug("Acquiring token...")
	result, err := ac.client.AcquireTokenSilent(ctx, ac.scopes)
	if err != nil {
		// cache miss, authenticate with another AcquireToken... method
		ac.logger.Debug("Cache miss, calling AcquireTokenByCredential...")

		result, err = ac.client.AcquireTokenByCredential(ctx, ac.scopes)
		if err != nil {
			return "", fmt.Errorf("error getting access token: %w", err)
		}
	}
	ac.logger.Debug("Successfully acquired token.")
	return AccessToken(result.AccessToken), nil
}
