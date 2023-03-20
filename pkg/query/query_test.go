package query

import (
	"context"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		q          store.Query
		objects    []models.Object
		rules      []models.AccessRule
		userGroups []string
		userID     string
		expected   []models.Object
	}{
		{
			name: "single accessible namespace + groups",
			q: &query{
				key:     "",
				value:   "",
				operand: OperandIncludes,
			},
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
			rules: []models.AccessRule{{
				Cluster:   "somecluster",
				Namespace: "ns-a",
				Subjects: []models.Subject{{
					Kind: "Role",
					Name: "some-role",
				}},
				AccessibleKinds: []string{"somekind"},
			}},
			userGroups: []string{"some-role"},
			expected: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
					Kind:      "somekind",
					Name:      "somename",
				},
			},
		},
		{
			name: "single accessible namespace + user",
			q: &query{
				key:     "",
				value:   "",
				operand: OperandIncludes,
			},
			userID: "some-user",
			rules: []models.AccessRule{{
				Cluster:   "somecluster",
				Namespace: "ns-a",
				Subjects: []models.Subject{{
					Kind: "Role",
					Name: "some-user",
				}},
				AccessibleKinds: []string{"somekind"},
			}},
			expected: []models.Object{{
				Cluster:   "somecluster",
				Namespace: "ns-a",
				Kind:      "somekind",
				Name:      "somename",
			}},
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &storefakes.FakeStoreReader{}
			reader.GetObjectsReturns(tt.objects, nil)
			reader.GetAccessRulesReturns(tt.rules, nil)

			qs, err := NewQueryService(context.Background(), QueryServiceOpts{
				Log:         logr.Discard(),
				StoreReader: reader,
			})

			assert.NoError(t, err)

			ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{
				ID:     tt.userID,
				Groups: tt.userGroups,
			})

			actual, err := qs.RunQuery(ctx, &query{})
			assert.NoError(t, err)

			diff := cmp.Diff(tt.expected, actual)

			if diff != "" {
				t.Errorf("RunQuery() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

type query struct {
	key     string
	value   string
	operand string
}

func (q *query) GetKey() string {
	return q.key
}

func (q *query) GetOperand() string {
	return q.operand
}

func (q *query) GetValue() string {
	return q.value
}

func (q *query) GetOffset() int64 {
	return 0
}

func (q *query) GetLimit() int64 {
	return 0
}
