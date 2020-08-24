package pages

import (
	"github.com/sclevine/agouti"
)

//ComponentsPage elements
type ComponentsPage struct {
	ClusterComponentsList *agouti.MultiSelection
}

//Dashboard initialises the page object
func Components(webDriver *agouti.Page) *ComponentsPage {
	componentsPage := ComponentsPage{
		ClusterComponentsList: webDriver.AllByXPath(`//*[@id="app"]/div/div[2]/div/div[3]/div/div[2]/div`),
	}

	return &componentsPage
}

func FindClusterComponent(componentsPage *ComponentsPage, componentName string) bool {

	count, _ := componentsPage.ClusterComponentsList.Count()
	for i := 0; i < count; i++ {

		text, _ := componentsPage.ClusterComponentsList.At(i).Text()

		if text == componentName {
			return true
		}
	}

	return false
}
