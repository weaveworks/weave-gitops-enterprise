package selectors

import (
	_ "embed"
	"log"
	"testing"

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

func Get(webDriver *agouti.Page, group, section, name string) *agouti.Selection {
	sel := selectorData[group][section][name]
	res := get(webDriver, sel)
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

func GetMulti(webDriver *agouti.Page, group, section, name string) *agouti.MultiSelection {
	sel := selectorData[group][section][name]
	res := getMulti(webDriver, sel)
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

func getMulti(webDriver *agouti.Page, selector map[string]string) *agouti.MultiSelection {
	selectAllByXPath := selector["selectAllByXPath"]
	if selectAllByXPath != "" {
		return webDriver.AllByXPath(selectAllByXPath)
	}
	selectAll := selector["selectAll"]
	if selectAll != "" {
		return webDriver.All(selectAll)
	}
	return nil
}

func get(webDriver *agouti.Page, selector map[string]string) *agouti.Selection {
	selectByXPath := selector["selectByXPath"]
	if selectByXPath != "" {
		return webDriver.FindByXPath(selectByXPath)
	}
	selectPattern := selector["select"]
	if selectPattern != "" {
		return webDriver.Find(selectPattern)
	}
	return nil
}
