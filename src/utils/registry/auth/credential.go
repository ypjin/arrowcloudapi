package auth

import (
	"net/http"
)

// Credential ...
type Credential interface {
	// AddAuthorization adds authorization information to request
	AddAuthorization(req *http.Request)
}

// Implements interface Credential
type basicAuthCredential struct {
	username string
	password string
}

// NewBasicAuthCredential ...
func NewBasicAuthCredential(username, password string) Credential {
	return &basicAuthCredential{
		username: username,
		password: password,
	}
}

func (b *basicAuthCredential) AddAuthorization(req *http.Request) {
	req.SetBasicAuth(b.username, b.password)
}

type cookieCredential struct {
	cookie *http.Cookie
}

// NewCookieCredential initialize a cookie based crendential handler, the cookie in parameter will be added to request to registry
// if this crendential is attached to a registry client.
func NewCookieCredential(c *http.Cookie) Credential {
	return &cookieCredential{
		cookie: c,
	}
}

func (c *cookieCredential) AddAuthorization(req *http.Request) {
	req.AddCookie(c.cookie)
}
