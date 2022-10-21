package estimation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
)

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
