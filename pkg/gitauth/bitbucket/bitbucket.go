package bitbucket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var scopes = []string{"REPO_WRITE", "REPO_READ", "PUBLIC_REPOS"}

type AuthClient interface {
	AuthURL(ctx context.Context, redirectURI string, state string) (url.URL, error)
	ExchangeCode(ctx context.Context, redirectURI, code string) (*TokenResponseState, error)
	ValidateToken(ctx context.Context, token string) error
}

func NewAuthClient(c *http.Client) AuthClient {
	return &defaultAuthClient{http: c}
}

type defaultAuthClient struct {
	http *http.Client
}

// AuthURL is used to construct the authorization URL.
// https://confluence.atlassian.com/bitbucketserver/bitbucket-oauth-2-0-provider-api-1108483661.html
func (c *defaultAuthClient) AuthURL(ctx context.Context, redirectURI string, state string) (url.URL, error) {
	u, err := buildBitbucketURL()

	if err != nil {
		return u, fmt.Errorf("cannot build bitbucket server base url: %w", err)
	}

	u.Path = "/rest/oauth2/latest/authorize"

	cid := getClientID()

	if cid == "" {
		return u, errors.New("env var BITBUCKET_SERVER_CLIENT_ID is not set")
	}

	params := u.Query()
	params.Set("client_id", cid)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	u.RawQuery = params.Encode()
	return u, nil
}

func (c *defaultAuthClient) ExchangeCode(ctx context.Context, redirectURI, code string) (*TokenResponseState, error) {
	u, err := buildBitbucketURL()

	if err != nil {
		return nil, fmt.Errorf("cannot build bitbucket server base url: %w", err)
	}

	cid := getClientID()
	if cid == "" {
		return nil, errors.New("env var BITBUCKET_SERVER_CLIENT_ID is not set")
	}

	secret := getClientSecret()
	if secret == "" {
		return nil, errors.New("env var BITBUCKET_SERVER_CLIENT_SECRET is not set")
	}
	// https://atlassian.example.com/rest/oauth2/latest/token?client_id=CLIENT_ID&client_secret=CLIENT_SECRET&code=CODE&grant_type=authorization_code&redirect_uri=REDIRECT_URI
	u.Path = "/rest/oauth2/latest/token"
	params := u.Query()
	params.Set("client_id", cid)
	params.Set("client_secret", secret)
	params.Set("redirect_uri", redirectURI)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	u.RawQuery = params.Encode()

	return doCodeExchangeRequest(ctx, u, c.http)
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

	return u, nil
}

func getClientID() string {
	return os.Getenv("BITBUCKET_SERVER_CLIENT_ID")
}

func getClientSecret() string {
	return os.Getenv("BITBUCKET_SERVER_CLIENT_SECRET")
}

func doCodeExchangeRequest(ctx context.Context, tURL url.URL, c *http.Client) (*TokenResponseState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create bitbucket code request: %w", err)
	}

	// Bitbucket requires this, else it will give a 400
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error exchanging bitbucket code: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		errRes := struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}{}

		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("could not parse error response: %w", err)
		}

		return nil, fmt.Errorf("code=%v, error=%s, description=%s", res.StatusCode, errRes.Error, errRes.Description)
	}

	r, err := parseTokenResponseBody(res.Body)
	if err != nil {
		return nil, err
	}

	token := &TokenResponseState{}

	token.SetTokenResponse(r)

	return token, nil
}

// TokenResponseState is used for passing state through HTTP middleware
type TokenResponseState struct {
	AccessToken    string
	TokenType      string
	ExpiresIn      time.Duration
	RefreshToken   string
	CreatedAt      int64
	HTTPStatusCode int
	Err            error
}

func (t *TokenResponseState) SetTokenResponse(token tokenRes) {
	t.AccessToken = token.AccessToken
	t.RefreshToken = token.RefreshToken
	t.ExpiresIn = time.Duration(token.ExpiresIn) * time.Second
	t.CreatedAt = token.CreatedAt
	t.TokenType = token.TokenType
}

type tokenRes struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int64  `json:"created_at"`
}

func parseTokenResponseBody(body io.ReadCloser) (tokenRes, error) {
	defer func() {
		_ = body.Close()
	}()

	var tokenResponse tokenRes
	err := json.NewDecoder(body).Decode(&tokenResponse)

	if err != nil {
		return tokenRes{}, err
	}

	return tokenResponse, nil
}
