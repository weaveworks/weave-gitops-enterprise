package pages

import (
	"fmt"
	"strings"

	"github.com/sclevine/agouti"
)

// Header webDriver elements
type TemplatesPage struct {
	TemplateHeader        *agouti.Selection
	TemplatesList         *agouti.MultiSelection
	TemplateProvider      *agouti.Selection
	TemplateProviderPopup *agouti.MultiSelection
	TemplateView          *agouti.MultiSelection
}

// TemplatesPage webdriver initialises the webDriver object
func GetTemplatesPage(webDriver *agouti.Page) *TemplatesPage {
	templatesPage := TemplatesPage{
		TemplateHeader:        webDriver.Find(`span[title="Templates"`),
		TemplatesList:         webDriver.All(`table tbody tr`),
		TemplateProvider:      webDriver.FindByID(`filter-by-provider`),
		TemplateProviderPopup: webDriver.All(`ul#filter-by-provider-popup li`),
		TemplateView:          webDriver.All(`#display-action > svg`),
	}

	return &templatesPage
}

type TemplateInformation struct {
	Name             string
	Type             *agouti.Selection
	Namespace        *agouti.Selection
	Provider         *agouti.Selection
	Description      *agouti.Selection
	CreateTemplate   *agouti.Selection
	ErrorHeader      *agouti.Selection
	ErrorDescription *agouti.Selection
}

func (t TemplatesPage) CountTemplateRows() int {
	count, _ := t.TemplatesList.Count()
	if count == 1 {
		msgTemplate, _ := t.TemplatesList.At(0).Text()
		if strings.Contains(msgTemplate, "No templates found") {
			return 0
		}
	}
	return count
}

func (t TemplatesPage) GetTemplateInformation(webDriver *agouti.Page, templateName string) *TemplateInformation {
	rowCount, _ := t.TemplatesList.Count()
	for i := 0; i < rowCount; i++ {
		row := t.TemplatesList.At(i).FindByXPath(fmt.Sprintf(`//td[1]//span[contains(text(), "%s")]/ancestor::tr`, templateName))
		if count, _ := row.Count(); count == 1 {
			return &TemplateInformation{
				Name:           templateName,
				Type:           row.FindByXPath(`td[2]`),
				Namespace:      row.FindByXPath(`td[3]`),
				Provider:       row.FindByXPath(`td[4]`),
				Description:    row.FindByXPath(`td[5]`),
				CreateTemplate: row.FindByXPath(`td[6]//button[@id="create-resource"]`),
			}
		}
	}
	return nil
}

func (t TemplatesPage) GetTemplateTableList() []string {
	rowCount, _ := t.TemplatesList.Count()
	rows := make([]string, rowCount)

	for i := 0; i < rowCount; i++ {
		rows[i], _ = t.TemplatesList.At(i).Text()
	}
	return rows
}
