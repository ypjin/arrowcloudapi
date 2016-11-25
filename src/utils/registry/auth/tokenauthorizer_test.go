package auth

import (
	"net/http"
	"testing"

	"utils/test"
)

func TestAuthorizeOfStandardTokenAuthorizer(t *testing.T) {
	handler := test.Handler(&test.Response{
		Body: []byte(`
		{
			"token":"token",
			"expires_in":300,
			"issued_at":"2016-08-17T23:17:58+08:00"
		}
		`),
	})

	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  "GET",
		Pattern: "/token",
		Handler: handler,
	})
	defer server.Close()

	authorizer := NewStandardTokenAuthorizer(nil, false, "repository", "library/ubuntu", "pull")
	req, err := http.NewRequest("GET", "http://registry", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	params := map[string]string{
		"realm": server.URL + "/token",
	}

	if err := authorizer.Authorize(req, params); err != nil {
		t.Fatalf("failed to authorize request: %v", err)
	}

	tk := req.Header.Get("Authorization")
	if tk != "Bearer token" {
		t.Errorf("unexpected token: %s != %s", tk, "Bearer token")
	}
}

func TestSchemeOfStandardTokenAuthorizer(t *testing.T) {
	authorizer := &standardTokenAuthorizer{}
	if authorizer.Scheme() != "bearer" {
		t.Errorf("unexpected scheme: %s != %s", authorizer.Scheme(), "bearer")
	}

}
