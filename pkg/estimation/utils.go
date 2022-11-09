package estimation

import (
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func mergeStringMaps(stringMaps ...map[string]string) map[string]string {
	cloned := map[string]string{}

	for _, update := range stringMaps {
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

// Takes a url in the format of aws://region?filters=filter1&filters=filter2
// returns the region and a map of filters
func ParseFilterURL(urlString string) (string, map[string]string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", nil, err
	}

	queryString := u.RawQuery
	if u.Scheme == "" {
		// urlString is just a query string
		queryString = urlString
		filters, err := ParseFilterQueryString(queryString)
		return "us-east-1", filters, err
	}

	if u.Scheme != "aws" {
		return "", nil, fmt.Errorf("invalid scheme %s, must be aws://", u.Scheme)
	}
	if u.Host == "" {
		return "", nil, errors.New("missing region must be aws://region")
	}

	filters, err := ParseFilterQueryString(queryString)
	return u.Host, filters, err
}

// ParseFilterQueryString parses a query string into a map of filters.
// Raises errors if duplicate or empty keys are provided.
func ParseFilterQueryString(annotations string) (map[string]string, error) {
	resultAnnotations := make(map[string]string)

	// check if annotation is a URL with a custom scheme or just as query string
	// isQueryString := true
	// if u, err := url.Parse(annotations); err == nil && u.Scheme != "" {
	// 	isQueryString = false
	// }

	parsedAnnot, err := url.ParseQuery(annotations)
	if err != nil {
		return nil, err
	}
	for k, v := range parsedAnnot {
		if len(v) > 1 {
			return nil, fmt.Errorf("annotation values cannot contain multiple values for the same key %s: %v", k, &v)
		}
		if len(v) < 1 || v[0] == "" {
			return nil, fmt.Errorf("invalid annotation values, cannot contain empty values %s: %v", k, annotations)

		}
		resultAnnotations[k] = v[0]

	}
	return resultAnnotations, nil
}
