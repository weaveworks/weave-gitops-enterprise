package estimation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
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

func Test_mergeStringMaps(t *testing.T) {
	mergeMapsTests := []struct {
		name     string
		origin   map[string]string
		updates  []map[string]string
		expected map[string]string
	}{
		{
			name: "updates overide origin",
			origin: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.medium",
			},
			updates: []map[string]string{
				{
					"operatingSystem": "Linux",
					"regionCode":      "us-iso-west-2",
					"instanceType":    "t3.large",
				},
			},
			expected: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-2",
				"instanceType":    "t3.large",
			},
		},
		{
			name: "updates with empty values are ignored",
			origin: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.medium",
			},
			updates: []map[string]string{
				{
					"operatingSystem": "",
					"regionCode":      "",
					"instanceType":    "t3.large",
				},
			},
			expected: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.large",
			},
		},
		{
			name: "Multiple updates provided to be merged with the origin",
			origin: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.medium",
			},
			updates: []map[string]string{
				{
					"operatingSystem": "Linux",
					"regionCode":      "us-iso-west-2",
					"instanceType":    "t3.medium",
				},
				{
					"operatingSystem": "Linux",
					"regionCode":      "us-iso-west-3",
					"instanceType":    "t3.large",
				},
			},
			expected: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-3",
				"instanceType":    "t3.large",
			},
		},
		{
			name: "Multiple updates provided to be merged with the origin, ignoring empty",
			origin: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-1",
				"instanceType":    "t3.medium",
			},
			updates: []map[string]string{
				{
					"operatingSystem": "Linux",
					"regionCode":      "us-iso-west-2",
					"instanceType":    "t3.large",
				},
				{
					"operatingSystem": "Linux",
					"regionCode":      "",
					"instanceType":    "",
				},
			},
			expected: map[string]string{
				"operatingSystem": "Linux",
				"regionCode":      "us-iso-west-2",
				"instanceType":    "t3.large",
			},
		},
	}

	for _, tt := range mergeMapsTests {
		t.Run(tt.name, func(t *testing.T) {
			mapsToMerge := append([]map[string]string{tt.origin}, tt.updates...)
			res := mergeStringMaps(mapsToMerge...)
			assert.Equal(t, res, tt.expected)

		})
	}
}

func Test_ParseFilterURL(t *testing.T) {
	// check that the region is parsed correctly
	region, _, err := ParseFilterURL("aws://us-east-1")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", region)

	// check that if the scheme is something other than aws we raise an error
	_, _, err = ParseFilterURL("gcp://us-east-1")
	assert.Error(t, err)

	// check that if the region is empty we raise an error
	_, _, err = ParseFilterURL("aws://")
	assert.Error(t, err)
	assert.Equal(t, "missing region must be aws://region", err.Error())

	// check that the filters are parsed correctly
	region, filters, err := ParseFilterURL("aws://us-east-1?foo=bar&baz=qux")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", region)
	assert.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, filters)

	// check that the filters are parsed correctly when there is no region
	_, _, err = ParseFilterURL("aws://?foo=bar&baz=qux")
	assert.Error(t, err)

	// if the scheme and host are omitted completely we should default to us-east-1
	region, filters, err = ParseFilterURL("foo=bar&baz=qux")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", region)
	assert.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, filters)

}

func Test_ParseFilterQueryString(t *testing.T) {
	ParseFilterQueryStringTests := []struct {
		name        string
		annotations string
		expected    map[string]string
		expectedErr string
	}{
		{
			name:        "Annotation string with single parameter",
			annotations: "instanceType=t3.large",
			expected: map[string]string{
				"instanceType": "t3.large",
			},
		},
		{
			name:        "Annotation string with multiple parameters",
			annotations: "instanceType=t3.large&regionCode=us-iso-east-1",
			expected: map[string]string{
				"instanceType": "t3.large",
				"regionCode":   "us-iso-east-1",
			},
		},
		{
			name:        "Annotation string with empty parameters",
			annotations: "",
			expected:    map[string]string{},
		},
		{
			name:        "Annotation string with multiple values for the same key",
			annotations: "instanceType=t3.large&instanceType=t3.medium",
			expected:    nil,
			expectedErr: "annotation values cannot contain multiple values for the same key",
		},
		{
			name:        "Annotation string with invalid query string",
			annotations: "instanceType!t3.large!regionCode!us-iso-east-1",
			expected:    nil,
			expectedErr: "invalid annotation values, cannot contain empty values",
		},
	}

	for _, tt := range ParseFilterQueryStringTests {
		t.Run(tt.name, func(t *testing.T) {

			res, err := ParseFilterQueryString(tt.annotations)
			if err != nil {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.Equal(t, res, tt.expected)
			}

		})
	}
}
