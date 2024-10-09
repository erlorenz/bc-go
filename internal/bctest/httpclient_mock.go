package bctest

import "net/http"

type MockTransport struct {
	Error    error
	Response *http.Response
}

func (mt MockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if mt.Error != nil {
		return nil, mt.Error
	}

	return mt.Response, nil
}
