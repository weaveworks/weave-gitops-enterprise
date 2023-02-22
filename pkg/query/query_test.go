package query

import (
	"context"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		q         Query
		objects   []models.Object
		rules     []models.AccessRule
		userRoles []string
		expected  []models.Object
	}{
		{
			name: "single accessible namespace",
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
				Cluster:         "somecluster",
				Namespace:       "ns-a",
				Principal:       "some-role",
				AccessibleKinds: []string{"somekind"},
			}},
			userRoles: []string{"some-role"},
			expected: []models.Object{
				{
					Cluster:   "somecluster",
					Namespace: "ns-a",
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
				Groups: tt.userRoles,
			})

			actual, err := qs.RunQuery(ctx, []Query{tt.q})
			assert.NoError(t, err)

			assert.EqualValues(t, tt.expected, actual)
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
