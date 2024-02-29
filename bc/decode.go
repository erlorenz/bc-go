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
	Error ErrorResponseError `json:"error"`
}

// The inner error field of the error response from BC.
type ErrorResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIError is a combination of the inner error and the StatusCode
// returned by the BC server when responding with an error status.
// It meets the Error interface.
type APIError struct {
	Code          string
	Message       string
	StatusCode    int
	CorrelationID GUID
	Request       *http.Request
}

func (err APIError) Error() string {
	return fmt.Sprintf("[%d %s] %q", err.StatusCode, err.Code, err.Message)
}

func newBCAPIError(statusCode int, code string, message string, request *http.Request) APIError {
	msg, id := extractCorrelationID(message)

	return APIError{
		Code:          code,
		Message:       msg,
		StatusCode:    statusCode,
		CorrelationID: id,
		Request:       request,
	}
}

// Decodes the http.Response into either an error or type T.
// The error can be inspected with errors.As to check if it is a
// APIError or an error during decoding.
func Decode[T Validator](r *http.Response) (T, error) {
	defer r.Body.Close()

	// Instantiate the generic data type early so it's zero
	// value can be returned if there is an error
	var data T

	// If error status call decodeErrorResponse() to return an error
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		err := decodeErrorResponse(r)
		return data, err
	}

	// Decode JSON into provided type if OK status
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return data, fmt.Errorf("could not decode %T: %w", data, err)
	}

	// Validate
	err = data.Validate()
	if err != nil {
		return data, fmt.Errorf("failed validation of %T: %w", data, err)
	}

	return data, nil

}

// Decodes the http.Response into an error.
// The error can be inspected with errors.As to check if it is a
// APIError or an error during decoding.
func DecodeNoContent(r *http.Response) error {
	defer r.Body.Close()

	// If error status call decodeErrorResponse() to return an error
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		err := decodeErrorResponse(r)
		return err
	}

	return nil

}

// MakeErrorFromResponse decodes the http.Response into an ErrorResponse struct
// and returns either an error with a failure to decode, or a BCServerError.
func decodeErrorResponse(r *http.Response) error {
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

	return newBCAPIError(r.StatusCode, data.Error.Code, data.Error.Message, r.Request)

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
