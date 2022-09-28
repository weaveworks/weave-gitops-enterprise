package selectors

import (
	_ "embed"
	"log"
	"testing"

	"github.com/osteele/liquid"
	"github.com/sclevine/agouti"
	"gopkg.in/yaml.v3"
)

//go:embed selectors.yaml
var SelectorsYaml string

type CommonSelectors map[string]map[string]map[string]map[string]string

func loadSelectors() CommonSelectors {
	data := SelectorsYaml
	s := CommonSelectors{}
	err := yaml.Unmarshal([]byte(data), &s)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	return s
}

var selectorData = loadSelectors()

var t *testing.T

func SetTestContext(test *testing.T) {
	t = test
}

func fromPairsToMap(pairs []string) map[string]interface{} {
	if len(pairs)%2 != 0 {
		t.Fatalf("key value pairs must be even")
	}
	m := make(map[string]interface{})
	for i := 0; i < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m
}

func substituteVariables(sel string, keyValuePairs []string) string {
	engine := liquid.NewEngine()
	bindings := fromPairsToMap(keyValuePairs)
	out, err := engine.ParseAndRenderString(sel, bindings)
	if err != nil {
		t.Fatalf("error parsing selector as a template: %v", err)
	}
	return out
}

func Get(webDriver *agouti.Page, group, section, name string, keyValuePairs ...string) *agouti.Selection {
	sel := selectorData[group][section][name]
	res := get(webDriver, sel, keyValuePairs...)
	if res == nil {
		t.Fatalf(
			"select or selectByXPath not found in selector: %v, at path: %v",
			sel,
			[]string{group, section, name},
		)
		return nil
	}
	return res
}

func GetMulti(webDriver *agouti.Page, group, section, name string, keyValuePairs ...string) *agouti.MultiSelection {
	sel := selectorData[group][section][name]
	res := getMulti(webDriver, sel, keyValuePairs...)
	if res == nil {
		t.Fatalf(
			"selectAll or selectAllByXPath not found in selector: %v, at path: %v",
			sel,
			[]string{group, section, name},
		)
		return nil
	}
	return res
}

func getMulti(webDriver *agouti.Page, selector map[string]string, keyValuePairs ...string) *agouti.MultiSelection {
	selectAllByXPath := selector["selectAllByXPath"]
	if selectAllByXPath != "" {
		return webDriver.AllByXPath(substituteVariables(selectAllByXPath, keyValuePairs))
	}
	selectAll := selector["selectAll"]
	if selectAll != "" {
		return webDriver.All(substituteVariables(selectAll, keyValuePairs))
	}
	return nil
}

func get(webDriver *agouti.Page, selector map[string]string, keyValuePairs ...string) *agouti.Selection {
	selectByXPath := selector["selectByXPath"]
	if selectByXPath != "" {
		return webDriver.FindByXPath(substituteVariables(selectByXPath, keyValuePairs))
	}
	selectPattern := selector["select"]
	if selectPattern != "" {
		return webDriver.Find(substituteVariables(selectPattern, keyValuePairs))
	}
	return nil
}
