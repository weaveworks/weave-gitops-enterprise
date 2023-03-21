package accesschecker

import (
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestHasAccess(t *testing.T) {

	tests := []struct {
		name    string
		objects []models.Object
		rules   []models.AccessRule
		user    *auth.UserPrincipal
		want    []bool
	}{
		{
			name: "no access",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "somekind",
					Name:      "somename",
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-b",
					Kind:      "somekind",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-b",
					}},
					AccessibleKinds: []string{"somekind"},
				},
			},
			want: []bool{false, false},
		},
		{
			name: "single accessible namespace + id",
			user: auth.NewUserPrincipal(auth.ID("some-admin"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "somekind",
					Name:      "somename",
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-b",
					Kind:      "somekind",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "User",
						Name: "some-admin",
					}},
					AccessibleKinds: []string{"somekind"},
				},
			},
			want: []bool{true, false},
		},
		{
			name: "single accessible namespace + groups",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "somekind",
					Name:      "somename",
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-b",
					Kind:      "somekind",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					AccessibleKinds: []string{"somekind"},
				},
			},
			want: []bool{true, false},
		},
		{
			name: "multiple clusters",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "cluster-a",
					Namespace: "ns-a",
					Kind:      "somekind",
					Name:      "somename",
				},
				{
					Cluster:   "cluster-b",
					Namespace: "ns-a",
					Kind:      "somekind",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "cluster-a",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					AccessibleKinds: []string{"somekind"},
				},
			},
			want: []bool{true, false},
		},
		{
			name: "inaccessible kinds",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "kind-a",
					Name:      "somename",
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "kind-b",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					AccessibleKinds: []string{"kind-a"},
				},
			},
			want: []bool{true, false},
		},
		{
			name: "multiple rules",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "kind-a",
					Name:      "somename",
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "kind-b",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					AccessibleKinds: []string{"kind-a"},
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					AccessibleKinds: []string{"kind-b"},
				},
			},
			want: []bool{true, true},
		},
		{
			name: "same kind inaccessible on a different clsuter",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "cluster-a",
					Namespace: "ns-a",
					Kind:      "kind-a",
					Name:      "somename",
				},
				{
					Cluster:   "cluster-b",
					Namespace: "ns-a",
					Kind:      "kind-a",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{{
				Cluster:   "cluster-a",
				Namespace: "ns-a",
				Subjects: []models.Subject{{
					Kind: "Group",
					Name: "group-a",
				}},
				AccessibleKinds: []string{"kind-a"},
			}},
			want: []bool{true, false},
		},
		{
			name: "wildcard kinds",
			user: auth.NewUserPrincipal(auth.ID("test"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "kind-a",
					Name:      "somename",
				},
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "kind-b",
					Name:      "somename",
				},
			},
			rules: []models.AccessRule{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					AccessibleKinds: []string{"*"},
				},
			},
			want: []bool{true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, obj := range tt.objects {
				ac := NewAccessChecker()

				got, err := ac.HasAccess(tt.user, obj, tt.rules)
				if err != nil {
					t.Errorf("accesschecker.HasAccess() error = %v, wantErr %v", err, false)
				}
				if got != tt.want[i] {
					t.Errorf("accesschecker.HasAccess() = %v, want %v", got, tt.want[i])
				}
			}
		})
	}

}
