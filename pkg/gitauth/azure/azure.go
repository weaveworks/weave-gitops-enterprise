package azure

import (
	"context"
	"encoding/base64"
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

// vso.code_write is used to read/write to Git repositories via the API.
// Using this scope also allows us to read account information which we
// use as a proxy for token validation.
var scopes = []string{"vso.code_write"}

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
// https://learn.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#2-authorize-your-app
func (c *defaultAuthClient) AuthURL(ctx context.Context, redirectURI string, state string) (url.URL, error) {
	u := buildAzureURL()
	u.Path = "/oauth2/authorize"

	id, err := getClientID()
	if err != nil {
		return u, err
	}

	params := u.Query()
	params.Set("client_id", id)
	params.Set("response_type", "Assertion")
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("redirect_uri", redirectURI)

	u.RawQuery = params.Encode()
	return u, nil
}

// ExchangeCode is called after the user authorizes the OAuth app to exchange a code for a token.
// https://learn.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#3-get-an-access-and-refresh-token-for-the-user
func (c *defaultAuthClient) ExchangeCode(ctx context.Context, redirectURI, code string) (*TokenResponseState, error) {
	u := buildAzureURL()
	u.Path = "/oauth2/token"

	secret, err := getClientSecret()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	params.Add("client_assertion", secret)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	params.Add("assertion", code)
	params.Add("redirect_uri", redirectURI)

	return doCodeExchangeRequest(ctx, u, c.http, strings.NewReader(params.Encode()))
}

// ValidateToken makes an HTTP call to https://app.vssps.visualstudio.com/_apis/accounts
// and returns a nil error if the response is 200 OK. Otherwise it returns an error.
// Making a call to get the accounts of the user is used as a proxy to validate whether
// the token is still valid.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/account/accounts/list?view=azure-devops-rest-7.0&tabs=HTTP
func (c *defaultAuthClient) ValidateToken(ctx context.Context, token string) error {
	u := buildAzureURL()
	u.Path = "_apis/accounts"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for Azure DevOps API: %w", err)
	}

	// https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?toc=%2Fazure%2Fdevops%2Forganizations%2Ftoc.json&view=azure-devops&tabs=Linux#use-a-pat
	b64Token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(":%s", token)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", b64Token))

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Azure DevOps API: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("token is invalid")
	}

	return nil
}

func buildAzureURL() url.URL {
	u := url.URL{}

	// Using a custom hostname here is purely for unit tests
	host, exists := os.LookupEnv("AZURE_DEVOPS_HOSTNAME")
	if !exists {
		host = "app.vssps.visualstudio.com"
	}
	u.Scheme = "https"
	u.Host = host

	return u
}

func getClientID() (string, error) {
	id, exists := os.LookupEnv("AZURE_DEVOPS_CLIENT_ID")
	if !exists {
		return "", errors.New("environment variable AZURE_DEVOPS_CLIENT_ID is not set")
	}

	return id, nil
}

func getClientSecret() (string, error) {
	secret, exists := os.LookupEnv("AZURE_DEVOPS_CLIENT_SECRET")
	if !exists {
		return "", errors.New("environment variable AZURE_DEVOPS_CLIENT_SECRET is not set")
	}

	return secret, nil
}

func doCodeExchangeRequest(ctx context.Context, tURL url.URL, c *http.Client, body io.Reader) (*TokenResponseState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("could not create azure code request: %w", err)
	}

	// POST request body is URL-encoded
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error exchanging azure code: %w", err)
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
	ExpiresIn    int64  `json:"expires_in,string"`
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
