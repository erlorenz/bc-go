package bc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

var fakeErrorResponse = ErrorResponse{
	Error: ODataError{"error_code", fmt.Sprintf("before_text  CorrelationId  %s.", validGUID)},
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

type fakeEntity struct {
	id        GUID
	quantity  int
	createdAt time.Time
	orderDate Date
}

func TestResponseData(t *testing.T) {
	fakeRecordJSON := []byte(fmt.Sprintf(`{"id": "%s","quantity":5,"createdAt":"2023-10-27T16:53:02.397Z","orderDate":"2024-02-20"}`, validGUID))
	body := io.NopCloser(bytes.NewBuffer(fakeRecordJSON))

	fakeResponse := &http.Response{StatusCode: 200, Body: body}

	record, err := Decode[fakeEntity](fakeResponse)

	if err != nil {
		t.Fatalf("failed to decode valid body: %s", err)
	}

	if record.orderDate.Day != 20 || record.orderDate.Month != 2 || record.orderDate.Year != 2024 {
		t.Errorf("date not valid: %v", record)
	}

	if record.quantity != 5 {
		t.Errorf("quantity not valid: %v", record)
	}

	if record.createdAt.IsZero() {
		t.Errorf("quantity not valid: %v", record)
	}

	if record.id == "" {
		t.Errorf("id not valid: %v", record)
	}
}
