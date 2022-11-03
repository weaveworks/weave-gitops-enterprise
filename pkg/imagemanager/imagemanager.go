package imagemanager

import (
	"context"
	"fmt"
	"time"

	automation "github.com/fluxcd/image-automation-controller/api/v1beta1"
	reflector "github.com/fluxcd/image-reflector-controller/api/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImageManager interface {
	AddImageAutomation(ctx context.Context, opts AddOptions, ref automation.CrossNamespaceSourceReference, policy reflector.ImagePolicyChoice, tags *reflector.TagFilter) error
}

type mgr struct {
	clients clustersmngr.ClustersManager
}

func NewImageManager(m clustersmngr.ClustersManager) ImageManager {
	return &mgr{
		clients: m,
	}
}

type AddOptions interface {
	GetName() string
	GetNamespace() string
	GetClusterName() string
	GetImage() string
	GetBranch() string
	GetPath() string
	GetSecretName() string
}

func (m *mgr) AddImageAutomation(ctx context.Context, opts AddOptions, ref automation.CrossNamespaceSourceReference, policy reflector.ImagePolicyChoice, tags *reflector.TagFilter) error {
	clients, err := m.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return fmt.Errorf("getting impersonated client: %w", err)
	}

	c, err := clients.Scoped(opts.GetClusterName())
	if err != nil {
		return fmt.Errorf("scoping client: %w", err)
	}

	r := reflector.ImageRepository{
		Spec: reflector.ImageRepositorySpec{
			Image:     opts.GetImage(),
			SecretRef: &meta.LocalObjectReference{Name: opts.GetSecretName()},
		},
	}
	r.Name = fmt.Sprintf("%s-repo", opts.GetName())
	r.Namespace = opts.GetNamespace()

	if err := c.Create(ctx, &r); err != nil {
		return fmt.Errorf("creating repository object: %w", err)
	}

	p := reflector.ImagePolicy{
		Spec: reflector.ImagePolicySpec{
			Policy:     policy,
			FilterTags: tags,
			ImageRepositoryRef: meta.NamespacedObjectReference{
				Name:      r.Name,
				Namespace: r.Namespace,
			},
		},
	}
	p.Name = fmt.Sprintf("%s-policy", opts.GetName())
	p.Namespace = opts.GetNamespace()

	if err := c.Create(ctx, &p); err != nil {
		return fmt.Errorf("creating policy object: %w", err)
	}

	a := automation.ImageUpdateAutomation{
		Spec: automation.ImageUpdateAutomationSpec{
			Interval:  v1.Duration{Duration: 1 * time.Minute},
			SourceRef: ref,
			GitSpec: &automation.GitSpec{
				Checkout: &automation.GitCheckoutSpec{
					Reference: v1beta2.GitRepositoryRef{
						Branch: opts.GetBranch(),
					},
				},
				Commit: automation.CommitSpec{
					Author: automation.CommitUser{
						Email: "fluxcdbot@users.noreply.github.com",
						Name:  "",
					},
					MessageTemplate: "{{range .Updated.Images}}{{println .}}{{end}}",
				},
				Push: &automation.PushSpec{
					Branch: opts.GetBranch(),
				},
			},
			Update: &automation.UpdateStrategy{
				Path:     opts.GetPath(),
				Strategy: automation.UpdateStrategySetters,
			},
		},
	}
	a.Name = fmt.Sprintf("%s-automation", opts.GetName())
	a.Namespace = opts.GetNamespace()

	if err := c.Create(ctx, &a); err != nil {
		return fmt.Errorf("creating automation object: %w", err)
	}

	return nil
}
