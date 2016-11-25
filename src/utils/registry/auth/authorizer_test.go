package auth

import (
	"net/http"
	"strings"
	"testing"

	"utils/test"

	"github.com/docker/distribution/registry/client/auth"
)

func TestNewAuthorizerStore(t *testing.T) {
	handler := test.Handler(&test.Response{
		StatusCode: http.StatusUnauthorized,
		Headers: map[string]string{
			"Www-Authenticate": "Bearer realm=\"https://auth.docker.io/token\",service=\"registry.docker.io\"",
		},
	})

	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  "GET",
		Pattern: "/v2/",
		Handler: handler,
	})
	defer server.Close()

	_, err := NewAuthorizerStore(server.URL, false, nil)
	if err != nil {
		t.Fatalf("failed to create authorizer store: %v", err)
	}
}

type simpleAuthorizer struct {
}

func (s *simpleAuthorizer) Scheme() string {
	return "bearer"
}

func (s *simpleAuthorizer) Authorize(req *http.Request,
	params map[string]string) error {
	req.Header.Set("Authorization", "Bearer token")
	return nil
}

func TestModify(t *testing.T) {
	authorizer := &simpleAuthorizer{}
	challenge := auth.Challenge{
		Scheme: "bearer",
	}

	as := &AuthorizerStore{
		authorizers: []Authorizer{authorizer},
		challenges:  []auth.Challenge{challenge},
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if err = as.Modify(req); err != nil {
		t.Fatalf("failed to modify request: %v", err)
	}

	header := req.Header.Get("Authorization")
	if len(header) == 0 {
		t.Fatal("\"Authorization\" header not found")
	}

	if !strings.HasPrefix(header, "Bearer") {
		t.Fatal("\"Authorization\" header does not start with \"Bearer\"")
	}
}
