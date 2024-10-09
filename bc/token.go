package bc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// Auth is used to retrieve an AccessToken.
// Implements the TokenGetter interface.
type Auth struct {
	client confidential.Client
	scopes []string
	logger *slog.Logger
}

// AccessToken is used in the Authorization header of requests.
type AccessToken string

// TokenGetter represents a client that retrieves
// an AccessToken.
type TokenGetter interface {
	GetToken(context.Context) (AccessToken, error)
}

// NewAuth validates the AuthParams and creates a new AuthClient.
func NewAuth(tenantID, clientID, clientSecret string) (*Auth, error) {

	cred, err := confidential.NewCredFromSecret(clientSecret)
	if err != nil {
		err = fmt.Errorf("authClient secret: %w", err)
		return nil, err
	}

	authority := "https://login.microsoft.com/" + string(tenantID)

	confidentialClient, err := confidential.New(authority, string(clientID), cred)
	if err != nil {
		err = fmt.Errorf("authclient confidentialClient: %w", err)
		return nil, err
	}

	// Don't think there is any reason to use a different one.
	// Can have this as a config param if ever need to.
	scopes := []string{"https://api.businesscentral.dynamics.com/.default"}

	return &Auth{
		client: confidentialClient,
		scopes: scopes,
		logger: slog.Default(),
	}, nil

}

func (ac *Auth) GetToken(ctx context.Context) (AccessToken, error) {

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
