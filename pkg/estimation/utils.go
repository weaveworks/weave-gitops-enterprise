package estimation

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func mergeStringMaps(origin map[string]string, updates ...map[string]string) map[string]string {
	cloned := map[string]string{}
	for k, v := range origin {
		cloned[k] = v
	}
	for _, update := range updates {
		for k, v := range update {
			if v != "" {
				cloned[k] = v
			}
		}
	}

	return cloned
}

func minMax(vals []float32) (float32, float32) {
	if len(vals) == 0 {
		return 0.0, 0.0
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })

	return vals[0], vals[len(vals)-1]
}

// reduceEstimates simplifies the estimate for provided clusters to a single
// Estimate.
//
// TODO: return an error if the currencies don't match across all Estimates?
func reduceEstimates(estimates map[string]*CostEstimate) *CostEstimate {
	lows, currency := func(m map[string]*CostEstimate) (float32, string) {
		currencies := sets.NewString()
		var total float32
		for _, v := range m {
			total += v.Low
			currencies.Insert(v.Currency)
		}
		return total, currencies.List()[0]
	}(estimates)
	highs := func(m map[string]*CostEstimate) float32 {
		var total float32
		for _, v := range m {
			total += v.High
		}
		return total
	}(estimates)

	return &CostEstimate{Low: lows, High: highs, Currency: currency}
}

func filtersFromMap(items map[string]string) []types.Filter {
	filters := []types.Filter{}
	for k, v := range items {
		filters = append(filters, types.Filter{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String(k),
			Value: aws.String(v),
		})
	}

	return filters
}
