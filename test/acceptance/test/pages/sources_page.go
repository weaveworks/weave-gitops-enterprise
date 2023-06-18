package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type SourcesPage struct {
	SourceHeader          *agouti.Selection
	ApplicationHeaderLink *agouti.Selection
	SourcesList           *agouti.Selection
	SupportEmailLink      *agouti.Selection
	MessageBar            *agouti.Selection
}

type SourceInformation struct {
	Name        *agouti.Selection
	Kind        *agouti.Selection
	Namespace   *agouti.Selection
	Tenant      *agouti.Selection
	Cluster     *agouti.Selection
	Status      *agouti.Selection
	Message     *agouti.Selection
	Url         *agouti.Selection
	Reference   *agouti.Selection
	Interval    *agouti.Selection
	LastUpdated *agouti.Selection
}

type SourceDetailPage struct {
	Header *agouti.Selection
	Title  *agouti.Selection
}

func GetSourcesPage(webDriver *agouti.Page) *SourcesPage {
	return &SourcesPage{
		SourceHeader:          webDriver.Find(`span[title="Sources"]`),
		ApplicationHeaderLink: webDriver.Find(`div[class*=Page__TopToolBar] a[href="/applications"]`),
		SourcesList:           webDriver.First(`table tbody`),
		SupportEmailLink:      webDriver.FindByLink(`support ticket`),
		MessageBar:            webDriver.FindByXPath(`//div[@id="root"]/div/main/div[2]`),
	}
}

func (a SourcesPage) FindSourceInList(sourceName string) *SourceInformation {
	source := a.SourcesList.FindByXPath(fmt.Sprintf(`//tr[.//a[.="%s"]]`, sourceName))
	return &SourceInformation{
		Name:        source.FindByXPath(`td[2]//a`),
		Kind:        source.FindByXPath(`td[3]`),
		Namespace:   source.FindByXPath(`td[4]`),
		Tenant:      source.FindByXPath(`td[5]`),
		Cluster:     source.FindByXPath(`td[6]`),
		Status:      source.FindByXPath(`td[7]`),
		Message:     source.FindByXPath(`td[8]`),
		Url:         source.FindByXPath(`td[9]`),
		Reference:   source.FindByXPath(`td[10]`),
		Interval:    source.FindByXPath(`td[11]`),
		LastUpdated: source.FindByXPath(`td[12]`),
	}
}

func (a SourcesPage) CountSources() int {
	sources := a.SourcesList.AllByXPath(`tr[.!="No data"]`)
	count, _ := sources.Count()
	return count
}

func GetSourceDetailPage(webDriver *agouti.Page) *SourceDetailPage {
	detailPage := SourceDetailPage{
		Header: webDriver.Find(`div[class*=Page__TopToolBar] span[class*=Breadcrumbs]`),
		Title:  webDriver.Find(`div[class*="SourceDetail"]`),
	}
	return &detailPage
}
