package registry

import (
	"fmt"
	"net/http"
	"testing"

	"utils/test"
)

type simpleModifier struct {
}

func (s *simpleModifier) Modify(req *http.Request) error {
	req.Header.Set("Authorization", "token")
	return nil
}

func TestRoundTrip(t *testing.T) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/",
			Handler: test.Handler(nil),
		})
	transport := NewTransport(&http.Transport{}, &simpleModifier{})
	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/", server.URL), nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("failed to send request: %s", err)
	}

	header := req.Header.Get("Authorization")
	if header != "token" {
		t.Errorf("unexpected header: %s != %s", header, "token")
	}

}
