package reconciler

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler interface {
	Setup(mgr ctrl.Manager) error
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}
