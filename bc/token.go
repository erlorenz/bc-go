package bc

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// AuthClient is used to retrieve an AccessToken.
// Implements the TokenGetter interface.
type AuthClient struct {
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

// NewAuthClient validates the AuthParams and creates a new AuthClient.
func NewAuthClient(tenantID GUID, clientID GUID, clientSecret string, logger *slog.Logger) (*AuthClient, error) {

	// Use default logger if none provided
	logger = cmp.Or(logger, slog.Default())

	// Validate config values
	if err := validateParams(tenantID, clientID, clientSecret); err != nil {
		err = fmt.Errorf("params validation error: %w", err)
		logger.Debug("validate params error", "error", err.Error())
		return nil, err
	}

	cred, err := confidential.NewCredFromSecret(clientSecret)
	if err != nil {
		err = fmt.Errorf("NewCredFromSecret error: %w", err)
		logger.Debug("NewCredFromSecret error", "error", err.Error())
		return nil, err
	}

	authority := "https://login.microsoft.com/" + string(tenantID)

	confidentialClient, err := confidential.New(authority, string(clientID), cred)
	if err != nil {
		err = fmt.Errorf("new confidentialClient error: %w", err)
		logger.Debug("new confidentialClient error", "error", err.Error())
		return nil, err
	}

	// Don't think there is any reason to use a different one.
	// Can have this as a config param if ever need to.
	scopes := []string{"https://api.businesscentral.dynamics.com/.default"}

	return &AuthClient{
		client: confidentialClient,
		scopes: scopes,
		logger: logger,
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

func validateParams(tenantID GUID, clientID GUID, clientSecret string) error {
	problems := []string{}
	if err := tenantID.Validate(); err != nil {
		problems = append(problems, fmt.Errorf("tenantID invalid (%w)", err).Error())
	}
	if err := clientID.Validate(); err != nil {
		problems = append(problems, fmt.Errorf("clientID invalid (%w)", err).Error())
	}
	if clientSecret == "" {
		problems = append(problems, "clientSecret invalid (is empty)")
	}
	if len(problems) > 0 {
		return fmt.Errorf(strings.Join(problems, ", "))
	}
	return nil
}
