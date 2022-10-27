package estimation

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MonthlyHours represents the number of hours AWS uses to calculate a price for
// a month.
const MonthlyHours = 730

// CostEstimate a price range for a cluster cost.
type CostEstimate struct {
	High     float32
	Low      float32
	Currency string
}

// Estimator implementations take a set of K8s resources, and estimate the cost
// for the items over a 730 hour period.
type Estimator interface {
	Estimate(context.Context, []*unstructured.Unstructured) (*CostEstimate, error)
}

// NilEstimator always returns a nil price.
func NilEstimator() Estimator {
	return nilEstimator{}
}

type nilEstimator struct{}

func (e nilEstimator) Estimate(context.Context, []*unstructured.Unstructured) (*CostEstimate, error) {
	return nil, nil
}
