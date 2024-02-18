package bcgo

import (
	"cmp"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// HTTPClient represents an HTTP client.
// It is usually an *http.Client from the standard library.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)
}

// ClientOption modifies the ClientOptions struct.
type ClientOptionFunc func(*ClientOptions)

// ClientOptions represents optional configuration values
// which are applied to the Client instead of using the defaults.
type ClientOptions struct {
	httpClient HTTPClient
	logger     *slog.Logger
}

// NewClientOptions uses the options provided and sets defaults
// for the rest.
func newClientOptions(optFuncs []ClientOptionFunc) *ClientOptions {
	options := &ClientOptions{}
	// Call each optFunc
	for _, optFunc := range optFuncs {
		optFunc(options)
	}

	// Set defaults
	options.httpClient = cmp.Or[HTTPClient](options.httpClient, &http.Client{Timeout: time.Second * 15})
	options.logger = cmp.Or(options.logger, slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	return options
}

// WithLogger sets a *slog.Logger instead of the default.
func WithLogger(logger *slog.Logger) ClientOptionFunc {
	return func(clientOptions *ClientOptions) {
		clientOptions.logger = logger
	}
}

// WithHTTPClient sets a client that meets the HTTPClient interface.
func WithHTTPClient(httpClient HTTPClient) ClientOptionFunc {
	return func(clientOptions *ClientOptions) {
		clientOptions.httpClient = httpClient
	}
}
