package registry

import (
	"net/http"

	"utils/log"
)

// Transport holds information about base transport and modifiers
type Transport struct {
	transport http.RoundTripper
	modifiers []Modifier
}

// NewTransport ...
func NewTransport(transport http.RoundTripper, modifiers ...Modifier) *Transport {
	return &Transport{
		transport: transport,
		modifiers: modifiers,
	}
}

// RoundTrip ...
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, modifier := range t.modifiers {
		if err := modifier.Modify(req); err != nil {
			return nil, err
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	log.Debugf("%d | %s %s", resp.StatusCode, req.Method, req.URL.String())

	return resp, err
}
