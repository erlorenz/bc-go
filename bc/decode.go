package bc

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// ErrorResponse is the body of the response returned from
// Business Central when the status is an error status.
type ErrorResponse struct {
	Error ODataError `json:"error"`
}

// ErrorResp is the contents of the error field. This is an
// OData compliant error.
type ODataError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BCServerError is a combination of the OData error and the StatusCode
// returned by the BC server when responding with an error status.
// It meets the Error interface.
type BCServerError struct {
	Code          string
	Message       string
	StatusCode    int
	CorrelationID GUID
}

func (err BCServerError) Error() string {
	return fmt.Sprintf("server responded with status %d: '%s', '%s'", err.StatusCode, err.Code, err.Message)
}

func newBCServerError(err ODataError, statusCode int) BCServerError {
	msg, id := extractCorrelationID(err.Message)

	return BCServerError{
		Code:          err.Code,
		Message:       msg,
		StatusCode:    statusCode,
		CorrelationID: id,
	}
}

// Decodes the http.Response into either an error or the type provided.
func Decode[T any](r *http.Response) (T, error) {

	var data T
	if r.StatusCode >= 200 && r.StatusCode < 300 {

		// Decode JSON into provided type if OK status
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return data, fmt.Errorf("could not decode %T: %w", data, err)
		}
		return data, nil
	}

	// If error status call makeFromErrorResponse() to return an error
	err := makeErrorFromResponse(r)

	return data, err

}

// MakeErrorFromResponse decodes the http.Response into an ErrorResponse struct
// and returns either an error with a failure to decode, or a BCServerError.
func makeErrorFromResponse(r *http.Response) error {
	defer r.Body.Close()
	var data ErrorResponse

	// Very strict response type, error on different structure.
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	err := d.Decode(&data)
	if err != nil {
		b, err := io.ReadAll(r.Body)
		slog.Default().Error(string(b))
		if err != nil {
			return fmt.Errorf("failed to read Response.Body: %s", err)
		}
		return fmt.Errorf("failed decoding Response.Body into ErrorResponse: %s", string(b))
	}

	return newBCServerError(data.Error, r.StatusCode)
}

// ExtractCorrelationID splits the message into a primary Message and then the CorrelationID.
// CorrelatioID may be an empty string as it does not always return one.
func extractCorrelationID(s string) (string, GUID) {
	// CorrelationID is surrounded by double spaces. ("  ")
	splits := strings.Split(s, "  ")

	msg := splits[0]

	// If it has CorrelationID split up
	if len(splits) == 3 {
		text := splits[2]
		// Has a trailing period ("."). Split at "." to avoid an error if this changes
		id := strings.Split(text, ".")[0]

		// Check that it is a GUID and if not just return the message
		if err := GUID(id).Validate(); err != nil {
			return msg, ""
		}

		return msg, GUID(id)
	}

	// No CorrelationID
	return msg, ""
}
