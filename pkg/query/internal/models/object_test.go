package models

import (
	"testing"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
)

func Test_IsExpired(t *testing.T) {
	//  Test that IsExpired returns true when the object is expired.

	tt := []struct {
		name     string
		policy   configuration.RetentionPolicy
		obj      Object
		expected bool
	}{
		{
			name:     "object is expired",
			policy:   configuration.RetentionPolicy(1 * time.Hour),
			obj:      Object{KubernetesDeletedAt: time.Now().Add(-2 * time.Hour)},
			expected: true,
		},
		{
			name:     "object is not expired",
			policy:   configuration.RetentionPolicy(1 * time.Hour),
			obj:      Object{KubernetesDeletedAt: time.Now().Add(-30 * time.Minute)},
			expected: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := IsExpired(tc.policy, tc.obj)
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
