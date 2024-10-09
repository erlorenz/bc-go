package bctest

import (
	"bytes"
	"encoding/json"
	"io"
)

// Utility to create a request body io.ReadCloser.
// Panics on error marshaling.
func NewRequestBody(v any) io.ReadCloser {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return io.NopCloser(bytes.NewBuffer(b))
}

const (
	// The index of the path segment "entitySetName" for custom endpoints
	PathIndexEntitySetName = 9
	// const pathIndexTenant = 2
	// const pathIndexEnvironment = 3
	// const pathIndexPublisher = 5
	// const pathIndexGroup = 6
	// const pathIndexVersion = 7
	// const pathIndexCompaniesSegment = 8

	// const pathIndexCommonVersion = 3
	// const pathIndexCommonCompaniesSegment = 6

	// The index of the path segment "entitySetName" for the common endpoint
	PathIndexCommonEntitySetName = 7
)
