package pages

import (
	"github.com/sclevine/agouti"
)

//ConfirmDisconnectClusterDialog  elements
type ConfirmDisconnectClusterDialog struct {
	AlertPopup             *agouti.Selection
	AlertDialogTitle       *agouti.Selection
	AlertDialogDescription *agouti.Selection
	ButtonRemove           *agouti.Selection
	ButtonCancel           *agouti.Selection
}

//GetConfirmDisconnectClusterDialog  initialises the webDriver object
func GetConfirmDisconnectClusterDialog(webDriver *agouti.Page) *ConfirmDisconnectClusterDialog {
	confirmDisconnectClusterDialog := ConfirmDisconnectClusterDialog{
		AlertPopup:             webDriver.Find(`#confirm-disconnect-cluster-dialog`),
		AlertDialogTitle:       webDriver.FindByXPath(`//*[@id="confirm-disconnect-cluster-dialog"]/*[@id="alert-dialog-title"]/h2`),
		AlertDialogDescription: webDriver.FindByXPath(`//*[@id="confirm-disconnect-cluster-dialog"]/*[@id="alert-dialog-description"]/p`),
		ButtonRemove:           webDriver.FindByXPath(`//*[@id="confirm-disconnect-cluster-dialog"]/div[3]/div/div[3]/button[1]`),
		ButtonCancel:           webDriver.FindByXPath(`//*[@id="confirm-disconnect-cluster-dialog"]/div[3]/div/div[3]/button[2]`),
	}

	return &confirmDisconnectClusterDialog
}
