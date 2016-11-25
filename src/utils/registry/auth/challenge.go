package auth

import (
	"net/http"

	au "github.com/docker/distribution/registry/client/auth"
)

// ParseChallengeFromResponse ...
func ParseChallengeFromResponse(resp *http.Response) []au.Challenge {
	challenges := au.ResponseChallenges(resp)

	return challenges
}
