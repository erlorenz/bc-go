package bcgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
)

var fakeErrorResponse = ErrorResponse{
	Error: ODataError{"error_code", fmt.Sprintf("before_text  CorrelationId  %s.", ValidGUID)},
}

var invalidErrorResponse = struct {
	Other string
}{
	Other: "other thing",
}

func newFakeBody(v any, t *testing.T) io.ReadCloser {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshall json in newFakeBody: %s", err)
	}
	return io.NopCloser(bytes.NewBuffer(b))
}

func TestResponseToError(t *testing.T) {

	fakeBody := newFakeBody(fakeErrorResponse, t)
	fakeResponse := &http.Response{StatusCode: 400, Body: fakeBody}

	err := makeErrorFromResponse(fakeResponse)

	var srvErr BCServerError

	if errors.As(err, &srvErr) {
		return
	}
	t.Errorf("failed making error from response: %s", err)

}

func TestResponseErrorResponseInvalid(t *testing.T) {

	fakeBody := newFakeBody(invalidErrorResponse, t)

	fakeResponse := &http.Response{StatusCode: 500, Body: fakeBody}

	err := makeErrorFromResponse(fakeResponse)

	var srvErr BCServerError

	if errors.As(err, &srvErr) {
		t.Errorf("returned Server error: %+v", err)
	}

	// t.Errorf("failed making error from response: %s", err)

}
