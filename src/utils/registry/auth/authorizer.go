package auth

import (
	"fmt"
	"net/http"
	"time"

	"utils"
	"utils/registry"

	au "github.com/docker/distribution/registry/client/auth"
)

// Authorizer authorizes requests according to the schema
type Authorizer interface {
	// Scheme : basic, bearer
	Scheme() string
	//Authorize adds basic auth or token auth to the header of request
	Authorize(req *http.Request, params map[string]string) error
}

// AuthorizerStore holds a authorizer list, which will authorize request.
// And it implements interface Modifier
type AuthorizerStore struct {
	authorizers []Authorizer
	challenges  []au.Challenge
}

// NewAuthorizerStore ...
func NewAuthorizerStore(endpoint string, insecure bool, authorizers ...Authorizer) (*AuthorizerStore, error) {
	endpoint = utils.FormatEndpoint(endpoint)

	client := &http.Client{
		Transport: registry.GetHTTPTransport(insecure),
		Timeout:   30 * time.Second,
	}

	resp, err := client.Get(buildPingURL(endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	challenges := ParseChallengeFromResponse(resp)
	return &AuthorizerStore{
		authorizers: authorizers,
		challenges:  challenges,
	}, nil
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}

// Modify adds authorization to the request
func (a *AuthorizerStore) Modify(req *http.Request) error {
	for _, challenge := range a.challenges {
		for _, authorizer := range a.authorizers {
			if authorizer.Scheme() == challenge.Scheme {
				if err := authorizer.Authorize(req, challenge.Parameters); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
