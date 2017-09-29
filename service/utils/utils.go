// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"net/http"
	"os"

	"arrowcloudapi/utils/log"
)

// VerifySecret verifies the UI_SECRET cookie in a http request.
func VerifySecret(r *http.Request) bool {
	secret := os.Getenv("UI_SECRET")
	c, err := r.Cookie("uisecret")
	if err != nil {
		log.Warningf("Failed to get secret cookie, error: %v", err)
	}
	return c != nil && c.Value == secret
}
