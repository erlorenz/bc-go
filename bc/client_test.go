package bc

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/internal/mock"
	"github.com/google/uuid"
)

// TestNewClient_Success tests successful client creation with minimal required parameters.
func TestNewClient_Success(t *testing.T) {
	t.Parallel()

	cred := &mock.Credential{}
	tenantID := uuid.NewString()
	companyID := uuid.NewString()

	client, err := NewClient(
		cred,
		tenantID,
		"Production",
		companyID,
		"v2.0",
		nil,
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("expected client, got nil")
	}

	// Verify internal state
	expectedBaseURL := fmt.Sprintf("https://api.businesscentral.dynamics.com/v2.0/%s/Production/api/v2.0/companies(%s)",
		tenantID, companyID)
	if client.getBaseURL() != expectedBaseURL {
		t.Errorf("expected baseURL %q, got %q", expectedBaseURL, client.getBaseURL())
	}

	if client.getRoute() != "v2.0" {
		t.Errorf("expected route %q, got %q", "v2.0", client.getRoute())
	}

	if !strings.HasPrefix(client.getUserAgent(), "bc-go/") {
		t.Errorf("expected user agent to start with 'bc-go/', got %q", client.getUserAgent())
	}

	if client.getHTTPClient().Timeout != 20*time.Second {
		t.Errorf("expected default timeout of 20s, got %v", client.getHTTPClient().Timeout)
	}
}

// TestNewClient_RequiredParameters tests validation of all required parameters.
func TestNewClient_RequiredParameters(t *testing.T) {
	t.Parallel()

	validCred := &mock.Credential{}
	validTenantID := uuid.NewString()
	validCompanyID := uuid.NewString()

	t.Run("nil credential", func(t *testing.T) {
		t.Parallel()
		_, err := NewClient(
			nil,
			validTenantID,
			"Production",
			validCompanyID,
			"v2.0",
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "cred cannot be nil") {
			t.Errorf("expected error about nil cred, got: %v", err)
		}
	})

	tests := []struct {
		name        string
		cred        *mock.Credential
		tenantID    string
		environment string
		companyID   string
		route       string
		wantErr     bool
	}{
		{
			name:        "empty tenantID",
			cred:        validCred,
			tenantID:    "",
			environment: "Production",
			companyID:   validCompanyID,
			route:       "v2.0",
			wantErr:     true,
		},
		{
			name:        "invalid tenantID format",
			cred:        validCred,
			tenantID:    "not-a-guid",
			environment: "Production",
			companyID:   validCompanyID,
			route:       "v2.0",
			wantErr:     true,
		},
		{
			name:        "empty environment",
			cred:        validCred,
			tenantID:    validTenantID,
			environment: "",
			companyID:   validCompanyID,
			route:       "v2.0",
			wantErr:     true,
		},
		{
			name:        "empty companyID",
			cred:        validCred,
			tenantID:    validTenantID,
			environment: "Production",
			companyID:   "",
			route:       "v2.0",
			wantErr:     true,
		},
		{
			name:        "invalid companyID format",
			cred:        validCred,
			tenantID:    validTenantID,
			environment: "Production",
			companyID:   "not-a-guid",
			route:       "v2.0",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewClient(
				tt.cred,
				tt.tenantID,
				tt.environment,
				tt.companyID,
				tt.route,
				nil,
			)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

// TestNewClient_RouteValidation tests route format validation.
func TestNewClient_RouteValidation(t *testing.T) {
	t.Parallel()

	cred := &mock.Credential{}
	tenantID := uuid.NewString()
	companyID := uuid.NewString()

	tests := []struct {
		name    string
		route   string
		wantErr bool
	}{
		{
			name:    "valid common route v2.0",
			route:   "v2.0",
			wantErr: false,
		},
		{
			name:    "valid extension route 3 segments",
			route:   "publisher/group/1.0",
			wantErr: false,
		},
		{
			name:    "invalid route - empty",
			route:   "",
			wantErr: true,
		},
		{
			name:    "invalid route - only 1 segment",
			route:   "single",
			wantErr: true,
		},
		{
			name:    "invalid route - only 2 segments",
			route:   "publisher/group",
			wantErr: true,
		},
		{
			name:    "invalid route - 4 segments",
			route:   "publisher/group/version/extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewClient(
				cred,
				tenantID,
				"Production",
				companyID,
				tt.route,
				nil,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// Verify route is set correctly
			if client.getRoute() != tt.route {
				t.Errorf("expected route %q, got %q", tt.route, client.getRoute())
			}

			// Verify baseURL contains the route
			if !strings.Contains(client.getBaseURL(), tt.route) {
				t.Errorf("expected baseURL to contain route %q, got %q", tt.route, client.getBaseURL())
			}
		})
	}
}

// TestNewClient_BaseURLConstruction tests that the base URL is constructed correctly.
func TestNewClient_BaseURLConstruction(t *testing.T) {
	t.Parallel()

	cred := &mock.Credential{}
	tenantID := uuid.NewString()
	companyID := uuid.NewString()

	tests := []struct {
		name        string
		environment string
		route       string
		rootURL     string
	}{
		{
			name:        "default root URL with v2.0",
			environment: "Production",
			route:       "v2.0",
			rootURL:     "",
		},
		{
			name:        "default root URL with extension",
			environment: "Sandbox",
			route:       "publisher/group/1.0",
			rootURL:     "",
		},
		{
			name:        "custom root URL",
			environment: "Production",
			route:       "v2.0",
			rootURL:     "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var opts *ClientOptions
			if tt.rootURL != "" {
				opts = &ClientOptions{RootURL: tt.rootURL}
			}

			client, err := NewClient(
				cred,
				tenantID,
				tt.environment,
				companyID,
				tt.route,
				opts,
			)

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// Construct expected base URL
			rootURL := tt.rootURL
			if rootURL == "" {
				rootURL = "https://api.businesscentral.dynamics.com"
			}
			expectedBaseURL := fmt.Sprintf("%s/v2.0/%s/%s/api/%s/companies(%s)",
				rootURL, tenantID, tt.environment, tt.route, companyID)

			if client.getBaseURL() != expectedBaseURL {
				t.Errorf("expected baseURL %q, got %q", expectedBaseURL, client.getBaseURL())
			}
		})
	}
}

// TestNewClient_Options tests that client options are applied correctly.
func TestNewClient_Options(t *testing.T) {
	t.Parallel()

	cred := &mock.Credential{}
	tenantID := uuid.NewString()
	companyID := uuid.NewString()

	t.Run("custom HTTPClient timeout", func(t *testing.T) {
		t.Parallel()

		customTimeout := 60 * time.Second
		customClient := &http.Client{
			Timeout: customTimeout,
		}

		client, err := NewClient(
			cred,
			tenantID,
			"Production",
			companyID,
			"v2.0",
			&ClientOptions{
				HTTPClient: customClient,
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if client.getHTTPClient().Timeout != customTimeout {
			t.Errorf("expected timeout %v, got %v", customTimeout, client.getHTTPClient().Timeout)
		}
	})

	t.Run("custom Logger", func(t *testing.T) {
		t.Parallel()

		customLogger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

		client, err := NewClient(
			cred,
			tenantID,
			"Production",
			companyID,
			"v2.0",
			&ClientOptions{
				Logger: customLogger,
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if client.getLogger() != customLogger {
			t.Error("expected custom logger to be preserved")
		}
	})

	t.Run("custom UserAgentSuffix", func(t *testing.T) {
		t.Parallel()

		suffix := "myapp/1.0.0"

		client, err := NewClient(
			cred,
			tenantID,
			"Production",
			companyID,
			"v2.0",
			&ClientOptions{
				UserAgentSuffix: suffix,
			},
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		userAgent := client.getUserAgent()
		expectedPrefix := "bc-go/"

		// Check prefix
		if !strings.HasPrefix(userAgent, expectedPrefix) {
			t.Errorf("expected user agent to start with %q, got %q", expectedPrefix, userAgent)
		}

		// Check suffix
		if !strings.HasSuffix(userAgent, suffix) {
			t.Errorf("expected user agent to end with %q, got %q", suffix, userAgent)
		}

		// Verify format is "bc-go/<version> <suffix>"
		expectedUA := fmt.Sprintf("bc-go/%s %s", Version, suffix)
		if userAgent != expectedUA {
			t.Errorf("expected user agent %q, got %q", expectedUA, userAgent)
		}
	})

	t.Run("nil options uses defaults", func(t *testing.T) {
		t.Parallel()

		client, err := NewClient(
			cred,
			tenantID,
			"Production",
			companyID,
			"v2.0",
			nil,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify defaults
		if client.getHTTPClient().Timeout != 20*time.Second {
			t.Errorf("expected default timeout of 20s, got %v", client.getHTTPClient().Timeout)
		}

		expectedUA := fmt.Sprintf("bc-go/%s", Version)
		if client.getUserAgent() != expectedUA {
			t.Errorf("expected user agent %q, got %q", expectedUA, client.getUserAgent())
		}

		if client.getLogger() == nil {
			t.Error("expected logger to be set")
		}
	})

	t.Run("empty options uses defaults", func(t *testing.T) {
		t.Parallel()

		client, err := NewClient(
			cred,
			tenantID,
			"Production",
			companyID,
			"v2.0",
			&ClientOptions{},
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify defaults
		if client.getHTTPClient().Timeout != 20*time.Second {
			t.Errorf("expected default timeout of 20s, got %v", client.getHTTPClient().Timeout)
		}

		expectedUA := fmt.Sprintf("bc-go/%s", Version)
		if client.getUserAgent() != expectedUA {
			t.Errorf("expected user agent %q, got %q", expectedUA, client.getUserAgent())
		}
	})
}
