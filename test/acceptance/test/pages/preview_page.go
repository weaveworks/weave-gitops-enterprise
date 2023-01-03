package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type Preview struct {
	Title    *agouti.Selection
	TabList  *agouti.Selection
	Path     *agouti.MultiSelection
	Text     *agouti.MultiSelection
	Download *agouti.Selection
	Close    *agouti.Selection
}

func GetPreview(webDriver *agouti.Page) Preview {
	return Preview{
		Title:    webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  h5`),
		TabList:  webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  div[role="tablist"]`),
		Path:     webDriver.All(`div[class*=MuiDialog-paper][role=dialog]  h6`),
		Text:     webDriver.All(`div[class*=MuiDialog-paper][role=dialog]  code`),
		Download: webDriver.Find(`div[class="info"] button`),
		Close:    webDriver.Find(`div[class*=MuiDialogTitle-root] button`),
	}
}

func (p Preview) GetPreviewTab(previewTab string) *agouti.Selection {
	return p.TabList.FindByXPath(fmt.Sprintf(`//button[.="%s"]`, previewTab))
}
