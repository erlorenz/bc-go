package bc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

var validErrorResponseJSON = []byte(fmt.Sprintf(`{
	"error": {
  		"code": "error_code",
  		"message": "before_text  CorrelationId  %s"
	}
}`, validGUID))

var invalidErrorResponseJSON = []byte(`{"other":"otherthing"}`)

func TestMakeErrorFromResponse(t *testing.T) {

	t.Logf("json: %s", validErrorResponseJSON)
	fakeResponse := &http.Response{StatusCode: 400, Body: newReadCloser(validErrorResponseJSON)}

	err := makeErrorFromResponse(fakeResponse)

	var srvErr BCServerError

	if errors.As(err, &srvErr) {
		return
	}
	t.Errorf("failed making error from response: %s", err)

}

func TestMakeErrorFromResponseInvalid(t *testing.T) {

	fakeResponse := &http.Response{StatusCode: 500, Body: newReadCloser(invalidErrorResponseJSON)}

	err := makeErrorFromResponse(fakeResponse)

	var srvErr BCServerError

	if errors.As(err, &srvErr) {
		t.Errorf("returned Server error: %+v", err)
	}

	// t.Errorf("failed making error from response: %s", err)

}

type fakeEntity struct {
	ID              GUID
	Quantity        int
	CreatedDateTime time.Time
	OrderDate       Date
	PostingDate     Date
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

	b, err := json.Marshal(fakeRecord)
	if err != nil {
		t.Fatalf("couldnt marshal: %s", err)
	}

	t.Logf("%s", b)

	fakeResponse := &http.Response{StatusCode: 200, Body: newReadCloser(b)}

	record, err := Decode[fakeEntity](fakeResponse)

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

	b, err := json.Marshal(fakeRecord)
	if err != nil {
		t.Fatalf("couldnt marshal: %s", err)
	}

	fakeResponse := &http.Response{StatusCode: 200, Body: newReadCloser(b)}

	record, err := Decode[fakeEntity](fakeResponse)

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
