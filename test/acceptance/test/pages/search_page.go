package pages

import (
	"fmt"
	"time"

	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/matchers"
)

type SearchPage struct {
	SearchBtn    *agouti.Selection
	Search       *agouti.Selection
	FilterBtn    *agouti.Selection
	FilterDialog *agouti.Selection
	ClearAllBtn  *agouti.Selection
}

func GetSearchPage(webDriver *agouti.Page) *SearchPage {
	return &SearchPage{
		SearchBtn:    webDriver.FindByXPath(`(//input[@placeholder="Search"]/ancestor::div[contains(@class, "TopBar")]/div[last()]//button)[1]`),
		Search:       webDriver.FindByID(`table-search`),
		FilterBtn:    webDriver.FindByXPath(`(//input[@placeholder="Search"]/ancestor::div[contains(@class, "TopBar")]/div[last()]//button)[2]`),
		FilterDialog: webDriver.Find(`div[class*="FilterDialog"].open`),
		ClearAllBtn:  webDriver.FindByXPath(`//div[.='Clear All']`),
	}
}

func (s SearchPage) SelectFilter(filterType string, filterID string, setCheck ...bool) {
	gomega.Eventually(s.FilterBtn.Click).Should(gomega.Succeed())
	gomega.Eventually(s.FilterDialog).Should(matchers.BeVisible())

	filters := s.FilterDialog.AllByXPath(`//form/ul/li`)
	fCount, _ := filters.Count()
	checkFilter := true
	if len(setCheck) > 0 {
		checkFilter = setCheck[0]
	}

	for i := 0; i < fCount; i++ {
		f := filters.At(i).FindByXPath(fmt.Sprintf(`//li/span[.="%s"]`, filterType))
		if count, _ := f.Count(); count == 1 {
			gomega.Eventually(func(g gomega.Gomega) {
				if checkFilter {
					g.Expect(filters.At(i).FindByXPath(fmt.Sprintf(`//input[@id="%s"]`, filterID)).Check()).Should(gomega.Succeed())
					g.Eventually(filters.At(i).FindByXPath(fmt.Sprintf(`//input[@id="%s"]/ancestor::span[contains(@class, "Mui-checked")]`, filterID)), time.Second*5).Should(matchers.BeFound())
				} else {
					g.Expect(filters.At(i).FindByXPath(fmt.Sprintf(`//input[@id="%s"]`, filterID)).Uncheck()).Should(gomega.Succeed())
					g.Eventually(filters.At(i).FindByXPath(fmt.Sprintf(`//input[@id="%s"]/ancestor::span[contains(@class, "Mui-checked")]`, filterID)), time.Second*5).ShouldNot(matchers.BeFound())
				}
			}, time.Second*30, time.Second*5).Should(gomega.Succeed(), fmt.Sprintf("Failed to select '%s' filter", filterID))
			break
		}
	}

	gomega.Expect(s.FilterBtn.Click()).Should(gomega.Succeed(), "Failed to close filter dialog")
}

func (s SearchPage) SearchName(searchWord string) {
	gomega.Eventually(s.SearchBtn.Click).Should(gomega.Succeed(), "Failed to click search buttton")
	gomega.Expect(s.Search.SendKeys(searchWord)).Should(gomega.Succeed(), "Failed to type violation name in search field")
	gomega.Expect(s.Search.SendKeys("\uE007")).Should(gomega.Succeed()) // send enter key code to do application search in table
}
