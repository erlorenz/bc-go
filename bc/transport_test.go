package bc

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erlorenz/bc-go/internal/mock"
)

// spyTransport is a test spy that captures the request it receives.
// It implements http.RoundTripper and records the request for inspection.
type spyTransport struct {
	capturedRequest *http.Request
	response        *http.Response
	err             error
}

func (s *spyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.capturedRequest = req
	if s.err != nil {
		return nil, s.err
	}
	if s.response != nil {
		return s.response, nil
	}
	// Return a basic 200 OK response by default
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}, nil
}

// TestBCTransport_AddsAuthorizationHeader verifies that the transport adds
// the Authorization header with the correct Bearer token format.
func TestBCTransport_AddsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	spy := &spyTransport{}
	cred := &mock.Credential{Token: "test-token-123"}
	userAgent := "bc-go/1.0.0"

	transport := newBCTransport(spy, cred, userAgent)

	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/test", nil)

	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spy.capturedRequest == nil {
		t.Fatal("expected request to be captured by spy, got nil")
	}

	authHeader := spy.capturedRequest.Header.Get("Authorization")
	expectedAuth := "Bearer test-token-123"

	if authHeader != expectedAuth {
		t.Errorf("expected Authorization header %q, got %q", expectedAuth, authHeader)
	}
}

// TestBCTransport_AddsUserAgentHeader verifies that the transport adds
// the User-Agent header with the correct value.
func TestBCTransport_AddsUserAgentHeader(t *testing.T) {
	t.Parallel()

	spy := &spyTransport{}
	cred := &mock.Credential{}
	userAgent := "bc-go/1.0.0 myapp/2.0"

	transport := newBCTransport(spy, cred, userAgent)

	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/test", nil)

	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spy.capturedRequest == nil {
		t.Fatal("expected request to be captured by spy, got nil")
	}

	uaHeader := spy.capturedRequest.Header.Get("User-Agent")

	if uaHeader != userAgent {
		t.Errorf("expected User-Agent header %q, got %q", userAgent, uaHeader)
	}
}

// TestBCTransport_AddsAcceptHeader verifies that the transport adds
// the Accept header with the correct OData no-metadata value.
func TestBCTransport_AddsAcceptHeader(t *testing.T) {
	t.Parallel()

	spy := &spyTransport{}
	cred := &mock.Credential{}
	userAgent := "bc-go/1.0.0"

	transport := newBCTransport(spy, cred, userAgent)

	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/test", nil)

	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spy.capturedRequest == nil {
		t.Fatal("expected request to be captured by spy, got nil")
	}

	acceptHeader := spy.capturedRequest.Header.Get("Accept")
	expectedAccept := "application/json;odata.metadata=none"

	if acceptHeader != expectedAccept {
		t.Errorf("expected Accept header %q, got %q", expectedAccept, acceptHeader)
	}
}

// TestBCTransport_ClonesRequest verifies that the transport clones the request
// and doesn't modify the original. This is important for retry logic.
func TestBCTransport_ClonesRequest(t *testing.T) {
	t.Parallel()

	spy := &spyTransport{}
	cred := &mock.Credential{}
	userAgent := "bc-go/1.0.0"

	transport := newBCTransport(spy, cred, userAgent)

	originalReq := httptest.NewRequest(http.MethodGet, "https://api.example.com/test", nil)

	// Store original header count
	originalHeaderCount := len(originalReq.Header)

	_, err := transport.RoundTrip(originalReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original request wasn't modified
	if len(originalReq.Header) != originalHeaderCount {
		t.Errorf("original request was modified: expected %d headers, got %d",
			originalHeaderCount, len(originalReq.Header))
	}

	// Original should not have the headers we added
	if originalReq.Header.Get("Authorization") != "" {
		t.Error("original request should not have Authorization header")
	}
	if originalReq.Header.Get("User-Agent") != "" {
		t.Error("original request should not have User-Agent header")
	}

	// But the captured request should have them
	if spy.capturedRequest.Header.Get("Authorization") == "" {
		t.Error("captured request should have Authorization header")
	}
	if spy.capturedRequest.Header.Get("User-Agent") == "" {
		t.Error("captured request should have User-Agent header")
	}
	if spy.capturedRequest.Header.Get("Accept") == "" {
		t.Error("captured request should have Accept header")
	}
}

// TestBCTransport_DelegatesToBaseTransport verifies that the transport
// calls the base transport's RoundTrip method.
func TestBCTransport_DelegatesToBaseTransport(t *testing.T) {
	t.Parallel()

	spy := &spyTransport{}
	cred := &mock.Credential{}
	userAgent := "bc-go/1.0.0"

	transport := newBCTransport(spy, cred, userAgent)

	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/test", nil)

	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spy.capturedRequest == nil {
		t.Error("base transport's RoundTrip was not called")
	}
}

// TestBCTransport_ErrorWhenGetTokenFails verifies that the transport returns
// an error when the credential fails to get a token.
func TestBCTransport_ErrorWhenGetTokenFails(t *testing.T) {
	t.Parallel()

	spy := &spyTransport{}
	expectedErr := errors.New("auth failed")
	cred := &mock.Credential{Err: expectedErr}
	userAgent := "bc-go/1.0.0"

	transport := newBCTransport(spy, cred, userAgent)

	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/test", nil)

	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("expected error when GetToken fails, got nil")
	}

	// Verify the error is wrapped properly
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap credential error, got: %v", err)
	}

	// Verify base transport was never called
	if spy.capturedRequest != nil {
		t.Error("base transport should not be called when GetToken fails")
	}
}

// TestBCTransport_UsesDefaultTransportWhenBaseIsNil verifies that
// newBCTransport uses http.DefaultTransport when base is nil.
func TestBCTransport_UsesDefaultTransportWhenBaseIsNil(t *testing.T) {
	t.Parallel()

	cred := &mock.Credential{}
	userAgent := "bc-go/1.0.0"

	transport := newBCTransport(nil, cred, userAgent)

	if transport.base != http.DefaultTransport {
		t.Error("expected base to be http.DefaultTransport when nil is passed")
	}
}
