package bcgo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type AuthClient struct {
	cc     confidential.Client
	scopes []string
	logger *slog.Logger
}

type AuthParams struct {
	TenantID     GUID
	ClientID     GUID
	ClientSecret string
	Scopes       []string
	Logger       *slog.Logger
}

func New(config AuthParams) (*AuthClient, error) {

	cred, err := confidential.NewCredFromSecret(config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("error creating MSAL confidential.Credential: %w", err)
	}

	authority := "https://login.microsoft.com/" + string(config.TenantID)

	confidentialClient, err := confidential.New(authority, string(config.ClientID), cred)
	if err != nil {
		return nil, fmt.Errorf("error creating MSAL confidential.Client: %w", err)
	}

	return &AuthClient{
		cc:     confidentialClient,
		scopes: config.Scopes,
		logger: config.Logger,
	}, nil

}

func (c *AuthClient) getToken(ctx context.Context) (string, error) {

	c.logger.Debug("Acquiring token...")
	result, err := c.cc.AcquireTokenSilent(ctx, c.scopes)
	if err != nil {
		// cache miss, authenticate with another AcquireToken... method
		c.logger.Debug("Cache miss, calling AcquireTokenByCredential...")

		result, err = c.cc.AcquireTokenByCredential(ctx, c.scopes)
		if err != nil {
			return "", fmt.Errorf("error getting access token: %w", err)
		}
	}
	c.logger.Debug("Successfully acquired token.")
	return result.AccessToken, nil
}
