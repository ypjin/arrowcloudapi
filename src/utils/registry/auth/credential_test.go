package auth

import (
	"net/http"
	"testing"
)

func TestAddAuthorizationOfBasicAuthCredential(t *testing.T) {
	cred := NewBasicAuthCredential("usr", "pwd")
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cred.AddAuthorization(req)

	usr, pwd, ok := req.BasicAuth()
	if !ok {
		t.Fatal("basic auth not found")
	}

	if usr != "usr" {
		t.Errorf("unexpected username: %s != usr", usr)
	}

	if pwd != "pwd" {
		t.Errorf("unexpected password: %s != pwd", pwd)
	}
}

func TestAddAuthorizationOfCookieCredential(t *testing.T) {
	cookie := &http.Cookie{
		Name:  "name",
		Value: "value",
	}
	cred := NewCookieCredential(cookie)
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cred.AddAuthorization(req)

	ck, err := req.Cookie("name")
	if err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	if ck.Value != "value" {
		t.Errorf("unexpected value: %s != value", ck.Value)
	}
}
