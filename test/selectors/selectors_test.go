package selectors

import (
	_ "embed"
	"testing"

	"github.com/sclevine/agouti"
	"github.com/stretchr/testify/assert"
)

var testData = map[string]map[string]map[string]map[string]string{
	"templates": {
		"page": {
			"singleHeader": {
				"select": "#header",
			},
			"singleClusterHeader": {
				"select": "#{{ cluster }} .header",
			},
			"multiRow": {
				"selectAll": "list-view tr",
			},
			"multiClusterRow": {
				"selectAll": "#{{ cluster }} list-view tr",
			},
		},
	},
}

func TestGetSelectors(t *testing.T) {
	SetTestContext(t)

	fakePage := agouti.JoinPage("localhost")

	single := get(fakePage, testData["templates"]["page"]["singleHeader"])
	assert.NotNil(t, single)

	multi := getMulti(fakePage, testData["templates"]["page"]["multiRow"])
	assert.NotNil(t, multi)

	// trying to get a single of a multi selector will nil out
	single = get(fakePage, testData["templates"]["page"]["multiRow"])
	assert.Nil(t, single)

	// trying to get a multi of a single selector will nil out
	multi = getMulti(fakePage, testData["templates"]["page"]["singleHeader"])
	assert.Nil(t, multi)
}

func TestTemplating(t *testing.T) {
	SetTestContext(t)

	fakePage := agouti.JoinPage("localhost")

	// Test template substitution
	singleTemplated := get(fakePage, testData["templates"]["page"]["singleClusterHeader"], "cluster", "dev")
	assert.Equal(t, "selection 'CSS: #dev .header [single]'", singleTemplated.String())

	// Test template substitution
	multiTemplated := getMulti(fakePage, testData["templates"]["page"]["multiClusterRow"], "cluster", "dev")
	assert.Equal(t, "selection 'CSS: #dev list-view tr'", multiTemplated.String())
}

func TestGetUnknownSelector(t *testing.T) {
	SetTestContext(t)

	fakePage := agouti.JoinPage("localhost")

	single := get(fakePage, testData["foo"]["bar"]["baz"])
	assert.Nil(t, single)

	multi := getMulti(fakePage, testData["foo"]["bar"]["baz"])
	assert.Nil(t, multi)
}
