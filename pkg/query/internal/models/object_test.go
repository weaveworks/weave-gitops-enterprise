package models

import (
	"testing"
	"time"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			name:     "object never expire if no policy",
			obj:      Object{KubernetesDeletedAt: time.Now().Add(-2 * time.Hour)},
			expected: false,
		},
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

func TestGetRelevantLabels(t *testing.T) {
	tests := []struct {
		name   string
		object defaultNormalizedObject
		want   map[string]string
	}{
		{
			name: "should return labels for kind with configured labels",
			object: defaultNormalizedObject{
				&gapiv1.GitOpsTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       gapiv1.Kind,
						APIVersion: "templates.weave.works/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-template-1",
						Namespace: "default",
						Labels: map[string]string{
							"weave.works/template-type": "cluster",
						},
					},
				},
				configuration.GitopsTemplateObjectKind,
			},
			want: map[string]string{
				"weave.works/template-type": "cluster",
			},
		},
		{
			name: "should return empty for kind with configured labels but object without labels",
			object: defaultNormalizedObject{
				&gapiv1.GitOpsTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       gapiv1.Kind,
						APIVersion: "templates.weave.works/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-template-1",
						Namespace: "default",
					},
				},
				configuration.GitopsTemplateObjectKind,
			},
			want: map[string]string{},
		},
		{
			name: "should return empty for kind without labels configured",
			object: defaultNormalizedObject{
				&sourcev1beta2.Bucket{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podinfo",
						Namespace: "test",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       sourcev1beta2.BucketKind,
						APIVersion: sourcev1beta2.GroupVersion.String(),
					},
					Spec: sourcev1beta2.BucketSpec{},
				},
				configuration.BucketObjectKind,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.object.GetRelevantLabels()); diff != "" {
				t.Fatalf("failed to get relevant label:\n%s", diff)
			}
		})
	}
}
