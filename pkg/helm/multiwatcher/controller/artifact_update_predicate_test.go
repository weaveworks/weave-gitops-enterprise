package controller

import (
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestArtifactUpdatePredicate_Update(t *testing.T) {
	tests := []struct {
		name  string
		event event.UpdateEvent
		want  bool
	}{
		{
			name:  "returns false no new or old object detected",
			event: event.UpdateEvent{},
			want:  false,
		},
		{
			name: "returns false if old source is not a sourcev1.Source object",
			event: event.UpdateEvent{
				ObjectOld: &corev1.Pod{},
				ObjectNew: &sourcev1beta2.HelmRepository{},
			},
			want: false,
		},
		{
			name: "returns false if new source is not a sourcev1.Source object",
			event: event.UpdateEvent{
				ObjectNew: &corev1.Pod{},
				ObjectOld: &sourcev1beta2.HelmRepository{},
			},
			want: false,
		},
		{
			name: "returns true if old source does not have an artifact but new does",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							Revision: "revision",
						},
					},
				},
				ObjectOld: &sourcev1beta2.HelmRepository{},
			},
			want: true,
		},
		{
			name: "returns true if old source and new source artifact revision doesn't match",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							Revision: "revision-2",
						},
					},
				},
				ObjectOld: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							Revision: "revision-1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false if old and new artifact revision are the same",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							Revision: "revision",
						},
					},
				},
				ObjectOld: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							Revision: "revision",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns true if old url and new url don't match",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							URL: "http://source-controller.flux-system.svc.cluster.local./gitrepository/default/go-demo-repo/40d6b21b888db0ca794876cf7bdd399e3da2137e.tar.gz",
						},
					},
				},
				ObjectOld: &sourcev1beta2.HelmRepository{
					Status: sourcev1beta2.HelmRepositoryStatus{
						Artifact: &sourcev1.Artifact{
							URL: "http://source-controller.demo-ns.svc.cluster.local./gitrepository/default/go-demo-repo/40d6b21b888db0ca794876cf7bdd399e3da2137e.tar.gz",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns true if the filter annotations don't match",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1beta2.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							HelmVersionFilterAnnotation: "> 0.0.0-0",
						},
					},
				},
				ObjectOld: &sourcev1beta2.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			he := ArtifactUpdatePredicate{}
			assert.Equalf(t, tt.want, he.Update(tt.event), "Update(\nold:\n%+v\nnew:\n%+v\n)", tt.event.ObjectOld, tt.event.ObjectNew)
		})
	}
}
