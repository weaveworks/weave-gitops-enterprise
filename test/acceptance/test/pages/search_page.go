package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

type SearchPage struct {
	SearchBtn    *agouti.Selection
	SearchForm   *agouti.Selection
	FilterBtn    *agouti.Selection
	FilterDialog *agouti.Selection
	Filters      *agouti.MultiSelection
}

func GetSearchPage(webDriver *agouti.Page) *SearchPage {
	return &SearchPage{
		SearchForm:   webDriver.FindByXPath(`//input[@placeholder="Search"]`),
		SearchBtn:    webDriver.AllByXPath(`//input[@placeholder="Search"]/ancestor::div/button`).At(0),
		FilterBtn:    webDriver.AllByXPath(`//input[@placeholder="Search"]/ancestor::div/button`).At(1),
		FilterDialog: webDriver.Find(`div[class*="FilterDialog"].open form`),
		Filters:      webDriver.AllByXPath(`//div[.="Filters"]/following-sibling::form/ul/li`),
	}
}

func (s SearchPage) GetFilter(filterType string, filterName string) *agouti.Selection {
	Eventually(s.FilterDialog).Should(BeVisible(), "Filter dialog can not be found")
	fCount, _ := s.Filters.Count()

	for i := 0; i < fCount; i++ {
		f := s.Filters.At(i).FindByXPath(fmt.Sprintf(`//li/span[.="%s"]`, filterType))
		if count, _ := f.Count(); count == 1 {
			filter := s.Filters.At(i).FindByXPath(fmt.Sprintf(`//li/span[.="%s"]/parent::li/div/div`, filterName))
			return filter
		}
	}
	return nil
}
