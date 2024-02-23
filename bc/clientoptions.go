package bc

import (
	"cmp"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type clientOptions struct {
	httpClient *http.Client
	logger     *slog.Logger
}

// ClientOption modifies the ClientOptions struct.
type ClientOptionFunc func(*clientOptions)

// clientOptions represents optional configuration values
// which are applied to the Client.

// newClientOptions uses the options provided and sets defaults
// for the rest.
func setClientOptions(c *Client, optFuncs []ClientOptionFunc) {
	options := clientOptions{}
	// Call each optFunc
	for _, optFunc := range optFuncs {
		optFunc(&options)
	}

	// Set defaults
	c.baseClient = cmp.Or(options.httpClient, &http.Client{Timeout: time.Second * 15})
	c.logger = cmp.Or(options.logger, slog.New(slog.NewJSONHandler(os.Stderr, nil)))

}

// WithLogger sets a *slog.Logger instead of the default.
func WithLogger(logger *slog.Logger) ClientOptionFunc {
	return func(clientOptions *clientOptions) {
		clientOptions.logger = logger
	}
}

// WithHTTPClient sets a client that meets the HTTPClient interface.
func WithHTTPClient(httpClient *http.Client) ClientOptionFunc {
	return func(clientOptions *clientOptions) {
		clientOptions.httpClient = httpClient
	}
}
