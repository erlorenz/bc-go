package bc

import (
	"log/slog"
	"net/http"
)

// ClientOption modifies the ClientOptions struct.
type ClientOption func(*Client)

// WithLogger sets a [slog.Logger] instead of the default.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(client *Client) {
		client.logger = logger
	}
}

// WithHTTPClient sets an http.Client instead of using the default baseClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.baseClient = httpClient
	}
}

// WithAuthClient sets a [TokenGetter] instead of constructing a default one.
// This is primarily for testing.
func WithAuthClient(authClient TokenGetter) ClientOption {
	return func(client *Client) {
		client.authClient = authClient
	}
}
