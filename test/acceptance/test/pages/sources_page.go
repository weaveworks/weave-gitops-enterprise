package pages

import (
	"github.com/sclevine/agouti"
)

type SourceDetailPage struct {
	Header *agouti.Selection
	Title  *agouti.Selection
}

func GetSourceDetailPage(webDriver *agouti.Page) *SourceDetailPage {
	detailPage := SourceDetailPage{
		Header: webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/applications"]/parent::node()/parent::node()/following-sibling::div[2]`),
		Title:  webDriver.Find(`div[class*="SourceDetail"]`),
	}
	return &detailPage
}
