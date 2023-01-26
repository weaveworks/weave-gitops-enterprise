package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type AuthClient interface {
	AuthURL(ctx context.Context, redirectURI string) (url.URL, error)
	ValidateToken(ctx context.Context, token string) error
}

func NewAuthClient(c *http.Client) AuthClient {
	return &defaultAuthClient{http: c}
}

type defaultAuthClient struct {
	http *http.Client
}

func (c *defaultAuthClient) AuthURL(ctx context.Context, redirectURI string) (url.URL, error) {
	u, err := buildBitbucketURL()

	if err != nil {
		return u, fmt.Errorf("building bitbucket server url: %w", err)
	}

	cid := getClientID()

	if cid == "" {
		return u, errors.New("env var BITBUCKET_SERVER_CLIENT_ID not set")
	}

	params := u.Query()
	params.Set("client_id", cid)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("grant_type", "authorization_code")

	// codeChallenge, err := verifier.CodeChallenge()
	// if err != nil {
	// 	return url.URL{}, fmt.Errorf("gitlab authorize url generate code challenge: %w", err)
	// }

	// params.Set("code_challenge", codeChallenge)
	// params.Set("code_challenge_method", "S256")
	// params.Set("scope", strings.Join(scopes, " "))
	u.RawQuery = params.Encode()
	return u, nil
}

func (c *defaultAuthClient) ValidateToken(ctx context.Context, token string) error {

	return nil
}

func buildBitbucketURL() (url.URL, error) {
	host := os.Getenv("BITBUCKET_SERVER_HOSTNAME")
	u := url.URL{}

	if host == "" {
		return u, errors.New("env var BITBUCKET_SERVER_HOSTNAME is not set")
	}

	u.Scheme = "https"
	u.Host = host

	u.Path = "/rest/oauth2/latest/authorize"

	return u, nil
}

func getClientID() string {
	return os.Getenv("BITBUCKET_SERVER_CLIENT_ID")
}

func getClientSecret() string {
	return os.Getenv("BITBUCKET_SERVER_CLIENT_SECRET")
}
