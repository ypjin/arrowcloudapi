package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// RequestHandlerMapping is a mapping between request and its handler
type RequestHandlerMapping struct {
	// Method is the method the request used
	Method string
	// Pattern is the pattern the request must match
	Pattern string
	// Handler is the handler which handles the request
	Handler func(http.ResponseWriter, *http.Request)
}

// ServeHTTP ...
func (rhm *RequestHandlerMapping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(rhm.Method) != 0 && r.Method != strings.ToUpper(rhm.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rhm.Handler(w, r)
}

// Response is a response used for unit test
type Response struct {
	// StatusCode is the status code of the response
	StatusCode int
	// Headers are the headers of the response
	Headers map[string]string
	// Boby is the body of the response
	Body []byte
}

// Handler returns a handler function which handle requst according to
// the response provided
func Handler(resp *Response) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if resp == nil {
			return
		}

		for k, v := range resp.Headers {
			w.Header().Add(http.CanonicalHeaderKey(k), v)
		}

		if resp.StatusCode == 0 {
			resp.StatusCode = http.StatusOK
		}
		w.WriteHeader(resp.StatusCode)

		if len(resp.Body) != 0 {
			io.Copy(w, bytes.NewReader(resp.Body))
		}
	}
}

// NewServer creates a HTTP server for unit test
func NewServer(mappings ...*RequestHandlerMapping) *httptest.Server {
	mux := http.NewServeMux()

	for _, mapping := range mappings {
		mux.Handle(mapping.Pattern, mapping)
	}

	return httptest.NewServer(mux)
}
