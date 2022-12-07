package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type Preview struct {
	Title    *agouti.Selection
	TabList  *agouti.Selection
	Text     *agouti.Selection
	Download *agouti.Selection
	Close    *agouti.Selection
}

func GetPreview(webDriver *agouti.Page) Preview {
	return Preview{
		Title:    webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  h5`),
		TabList:  webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  div[role="tablist"]`),
		Text:     webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  code`),
		Download: webDriver.Find(`div[class="info"] button`),
		Close:    webDriver.Find(`div[class*=MuiDialogTitle-root] button`),
	}
}

func (p Preview) GetPreviewTab(previewTab string) *agouti.Selection {
	return p.TabList.FindByXPath(fmt.Sprintf(`//button[.="%s"]`, previewTab))
}
