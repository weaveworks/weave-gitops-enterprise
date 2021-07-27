package clusters_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/cmd/mccp/pkg/clusters"
)

func TestListClusters(t *testing.T) {
	tests := []struct {
		name             string
		cs               []clusters.Cluster
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no clusters",
			expected: "No clusters found.\n",
		},
		{
			name: "clusters exist",
			cs: []clusters.Cluster{
				{
					Name:   "cluster-a",
					Status: "status-a",
				},
				{
					Name:   "cluster-b",
					Status: "status-b",
				},
			},
			expected: "NAME\tSTATUS\ncluster-a\tstatus-a\ncluster-b\tstatus-b\n",
		},
		{
			name:             "error retrieving clusters",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve clusters from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(tt.cs, "", tt.err)
			w := new(bytes.Buffer)
			err := clusters.ListClusters(c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeClient struct {
	cs  []clusters.Cluster
	s   string
	err error
}

func NewFakeClient(cs []clusters.Cluster, s string, err error) *FakeClient {
	return &FakeClient{
		cs:  cs,
		s:   s,
		err: err,
	}
}

func (c *FakeClient) Source() string {
	return "In-memory fake"
}

func (c *FakeClient) RetrieveClusters() ([]clusters.Cluster, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.cs, nil
}

func (c *FakeClient) GetClusterKubeconfig(name string) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}
