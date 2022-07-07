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
		FilterDialog: webDriver.Find(`div[class*="FilterDialog"].open`),
	}
}

func (s SearchPage) SelectFilter(filterType string, filterID string) {
	Eventually(s.FilterDialog).Should(BeVisible(), "Filter dialog can not be found")
	filters := s.FilterDialog.AllByXPath(`//form/ul/li`)
	fCount, _ := filters.Count()

	for i := 0; i < fCount; i++ {
		f := filters.At(i).FindByXPath(fmt.Sprintf(`//li/span[.="%s"]`, filterType))
		if count, _ := f.Count(); count == 1 {

			Eventually(func(g Gomega) {
				g.Expect(filters.At(i).FindByXPath(fmt.Sprintf(`//input[@id="%s"]`, filterID)).Check()).Should(Succeed())
				g.Expect(filters.At(i).FindByXPath(fmt.Sprintf(`//input[@id="%s"]/ancestor::span[contains(@class, "Mui-checked")]`, filterID))).Should(BeFound())
			}).Should(Succeed(), "Failed to select cluster filter: "+filterID)

		}
	}

}
