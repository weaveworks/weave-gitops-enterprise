package pages

import (
	"github.com/sclevine/agouti"
)

type SourceDetailPage struct {
	Header *agouti.Selection
	Title  *agouti.Selection
}

func GetSourceDetailPage(webDriver *agouti.Page) *ApplicationDetailPage {
	detailPage := ApplicationDetailPage{
		Header: webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/applications"]/parent::node()/parent::node()/following-sibling::div[2]`),
		Title:  webDriver.Find(`[class*=DetailTitle]`),
	}
	return &detailPage
}
