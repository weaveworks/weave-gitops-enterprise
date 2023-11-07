package git

import (
	"fmt"

	"github.com/go-logr/logr"
)

// ProviderCreator defines the interface for creating a Git provider.
type ProviderCreator interface {
	Create(providerName string, opts ...ProviderWithFn) (Provider, error)
}

// ProviderFactory is used to create and return a
// concrete git provider.
type ProviderFactory struct {
	log logr.Logger
}

// NewFactory creates a new factory for git providers.
func NewFactory(log logr.Logger) *ProviderFactory {
	return &ProviderFactory{
		log: log,
	}
}

// Create creates and returns a new git provider.
func (f *ProviderFactory) Create(providerName string, opts ...ProviderWithFn) (Provider, error) {
	var (
		provider Provider
		err      error
	)

	switch providerName {
	case GitHubProviderName:
		provider, err = NewGitHubProvider(f.log)
	case GitLabProviderName:
		provider, err = NewGitLabProvider(f.log)
	case BitBucketServerProviderName:
		provider, err = NewBitBucketServerProvider(f.log)
	case AzureDevOpsProviderName:
		provider, err = NewAzureDevOpsProvider(f.log)
	default:
		return nil, fmt.Errorf("provider %q is not supported", providerName)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to use provider %q: %w", providerName, err)
	}

	option := ProviderOption{}

	for _, opt := range opts {
		if err = opt(&option); err != nil {
			return nil, fmt.Errorf("unable to gather options on provider %q: %w", providerName, err)
		}
	}

	if err := provider.Setup(option); err != nil {
		return nil, fmt.Errorf("unable to apply options on provider %q: %w", providerName, err)
	}

	return provider, nil
}

type ProviderOption struct {
	Hostname            string
	OAuth2Token         string
	TokenType           string
	Token               string
	Username            string
	ConditionalRequests bool
}

type ProviderWithFn func(o *ProviderOption) error

func WithDomain(domain string) ProviderWithFn {
	return func(p *ProviderOption) error {
		p.Hostname = addSchemeToDomain(domain)

		return nil
	}
}

func WithOAuth2Token(token string) ProviderWithFn {
	return func(p *ProviderOption) error {
		p.OAuth2Token = token

		return nil
	}
}

func WithToken(tokenType, token string) ProviderWithFn {
	return func(p *ProviderOption) error {
		p.TokenType = tokenType
		p.Token = token

		return nil
	}
}

func WithConditionalRequests() ProviderWithFn {
	return func(p *ProviderOption) error {
		p.ConditionalRequests = true

		return nil
	}
}

func WithoutConditionalRequests() ProviderWithFn {
	return func(p *ProviderOption) error {
		p.ConditionalRequests = false

		return nil
	}
}

func WithUsername(username string) ProviderWithFn {
	return func(p *ProviderOption) error {
		p.Username = username

		return nil
	}
}
