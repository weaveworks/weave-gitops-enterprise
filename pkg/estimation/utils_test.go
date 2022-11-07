package estimation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_reduceClusterEstimates(t *testing.T) {
	estimates := map[string]*CostEstimate{
		"cluster1": {Low: 10, High: 50, Currency: "USD"},
		"cluster2": {Low: 5, High: 45, Currency: "USD"},
		"cluster3": {Low: 20, High: 80, Currency: "USD"},
	}

	reduced := reduceEstimates(estimates)

	// Min = 5 + 10 + 20 = 35
	// Max = 50 + 45 + 80 = 175
	want := &CostEstimate{Low: 35, High: 175, Currency: "USD"}
	if diff := cmp.Diff(want, reduced); diff != "" {
		t.Fatalf("failed to reduce:\n%s", diff)
	}
}
