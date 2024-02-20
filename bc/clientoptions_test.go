package bc

import (
	"log/slog"
	"net/http"
	"testing"
)

func TestNewClientOptions(t *testing.T) {

	allOptFuncs := []ClientOptionFunc{WithLogger(slog.Default()), WithHTTPClient(http.DefaultClient)}
	noOptFuncs := []ClientOptionFunc{}

	t.Run("AllNewClientOptions", func(t *testing.T) {
		options := newClientOptions(allOptFuncs)
		if options.logger == nil {
			t.Error("logger is nil")
		}

		if options.httpClient == nil {
			t.Error("httpClient is nil")
		}
	})
	t.Run("NoNewClientOptions", func(t *testing.T) {
		options := newClientOptions(noOptFuncs)
		if options.logger == nil {
			t.Error("logger is nil")
		}

		if options.httpClient == nil {
			t.Error("httpClient is nil")
		}
	})

}
