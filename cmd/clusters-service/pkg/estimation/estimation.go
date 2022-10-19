package estimation

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// CostEstimate a price range for a cluster cost.
type CostEstimate struct {
	High     float32
	Low      float32
	Currency string
}

// Estimator implementations take a set of K8s resources, and estimate the cost
// for the items.
type Estimator interface {
	Estimate([]*unstructured.Unstructured) (*CostEstimate, error)
}

// NilEstimator always returns a nil price.
func NilEstimator() Estimator {
	return nilEstimator{}
}

type nilEstimator struct{}

func (e nilEstimator) Estimate([]*unstructured.Unstructured) (*CostEstimate, error) {
	return nil, nil
}
