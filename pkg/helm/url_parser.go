package helm

import (
	"fmt"
	"net/url"
	"strings"
)

// Service represents the elements that we need to use the client Proxy to fetch
// a URL.
type Service struct {
	Scheme    string
	Namespace string
	Name      string
	Path      string
	Port      string
}

// ParseArtifactURL takes HelmRepository Artifact URL for a remote cluster and
// returns the components of the URL.
func ParseArtifactURL(artifactURL string) (*Service, error) {
	u, err := url.Parse(artifactURL)
	if err != nil {
		return nil, err
	}

	// Split hostname to get namespace and name.
	host := strings.Split(u.Hostname(), ".")

	if len(host) != 6 || host[2] != "svc" || u.Path == "/" {
		return nil, fmt.Errorf("invalid artifact URL %s", artifactURL)
	}

	port := u.Port()
	if port == "" {
		port = "80"
	}

	// When we use Helm to fetch the index file, it appends "/index.yaml" to the
	// artifact URL which causes it to 404 so this is trimmed.
	if strings.HasSuffix(u.Path, ".yaml/index.yaml") {
		u.Path = strings.TrimSuffix(u.Path, "/index.yaml")
	}

	return &Service{
		Scheme:    u.Scheme,
		Namespace: host[1],
		Name:      host[0],
		Path:      u.Path,
		Port:      port,
	}, nil
}
