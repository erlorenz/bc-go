package bc_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
	"github.com/erlorenz/bc-go/bctest"
)

var validGUID = bctest.NewGUID()

var validErrorResponse = bc.ErrorResponse{
	Error: bc.ErrorResponseError{
		Code:    "error_code",
		Message: fmt.Sprintf("before_text  CorrelationId  %s", validGUID),
	},
}

var invalidErrorResponse = struct {
	OtherField string `json:"otherField"`
}{
	OtherField: "something else",
}

func TestMakeErrorFromResponse(t *testing.T) {

	body := bctest.NewRequestBody(validErrorResponse)
	fakeResponse := &http.Response{StatusCode: 400, Body: body}

	_, err := bc.Decode[bc.Validator](fakeResponse)
	if err == nil {
		t.Logf("response status: %d", fakeResponse.StatusCode)
		b, _ := io.ReadAll(fakeResponse.Body)
		t.Logf("response body: %s", string(b))
		t.Fatal("expected error, got nil")
	}

	var srvErr bc.APIError
	if errors.As(err, &srvErr) {
		return
	}
	t.Errorf("failed making error from response: %s", err)

}

func TestMakeErrorFromResponseInvalid(t *testing.T) {

	body := bctest.NewRequestBody(invalidErrorResponse)
	fakeResponse := &http.Response{StatusCode: 400, Body: body}

	_, err := bc.Decode[bc.Validator](fakeResponse)
	if err == nil {
		t.Logf("response status: %d", fakeResponse.StatusCode)
		b, _ := io.ReadAll(fakeResponse.Body)
		t.Logf("response body: %s", string(b))
		t.Fatal("expected error, got nil")
	}

	var srvErr bc.APIError
	if errors.As(err, &srvErr) {
		t.Logf("error: %s", srvErr)
		t.Errorf("was read as a ServerError, it should not be.")
	}

}

type fakeEntity struct {
	ID              bc.GUID
	Quantity        int
	CreatedDateTime time.Time
	OrderDate       bc.Date
	PostingDate     bc.Date
	Number          string
}

func (f fakeEntity) Validate() error {
	if f.ID == "" {
		return errors.New("validation error: ID is empty")
	}
	return nil
}

func TestResponseData(t *testing.T) {
	fakeRecord := map[string]any{
		"ID":              validGUID,
		"Quantity":        5,
		"CreatedDateTime": "2024-02-20T00:12:20.273Z",
		"OrderDate":       "2024-02-20",
		"PostingDate":     "2024-02-20",
		"Number":          "NUMBER324234",
	}

	fakeResponse := &http.Response{StatusCode: 200, Body: bctest.NewRequestBody(fakeRecord)}

	record, err := bc.Decode[fakeEntity](fakeResponse)

	t.Logf("%v", record)

	if err != nil {
		t.Fatalf("failed to decode valid body: %s", err)
	}

	if record.OrderDate.Day != 20 || record.OrderDate.Month != 2 || record.OrderDate.Year != 2024 {
		t.Errorf("date not valid: %v", record)
	}

	if record.Quantity != 5 {
		t.Errorf("quantity not valid: %v", record)
	}

	if record.CreatedDateTime.IsZero() {
		t.Errorf("quantity not valid: %v", record)
	}

	if record.ID == "" {
		t.Errorf("id not valid: %v", record)
	}

	if record.Number == "" {
		t.Errorf("number not valid: %v", record)
	}
}

func TestResponseDataInvalid(t *testing.T) {
	fakeRecord := map[string]any{
		"ID":              validGUID,
		"Quantity":        5,
		"CreatedDateTime": time.Now(),
		"PostingDate":     "2024-02-20",
		"OrderDate":       "2024-02-20",
		"Number":          "NUMBER324234",
	}

	body := bctest.NewRequestBody(fakeRecord)

	fakeResponse := &http.Response{StatusCode: 200, Body: body}

	record, err := bc.Decode[fakeEntity](fakeResponse)

	if err != nil {
		t.Fatalf("failed to decode valid body: %s", err)
	}

	t.Logf("%+#v", record)

	if record.OrderDate.Day != 20 || record.OrderDate.Month != 2 || record.OrderDate.Year != 2024 {
		t.Errorf("date not valid: %v", record)
	}

	if record.Quantity != 5 {
		t.Errorf("quantity not valid: %v", record)
	}

	if record.CreatedDateTime.IsZero() {
		t.Errorf("quantity not valid: %v", record)
	}

	if record.ID == "" {
		t.Errorf("id not valid: %v", record)
	}

	if record.Number == "" {
		t.Errorf("number not valid: %v", record)
	}
}
