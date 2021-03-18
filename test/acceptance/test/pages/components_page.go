package pages

import (
	"github.com/sclevine/agouti"
)

//ComponentsPage elements
type ComponentsPage struct {
	ClusterComponentsList *agouti.MultiSelection
}

type ClusterComponent struct {
	Name       string
	StatusNode *agouti.Selection
}

//Dashboard initialises the page object
func Components(webDriver *agouti.Page) *ComponentsPage {
	componentsPage := ComponentsPage{
		ClusterComponentsList: webDriver.AllByXPath(`//*[@id="app"]/div/div[2]/div/div[3]/div/div[2]/div`),
	}

	return &componentsPage
}

func FindClusterComponent(componentsPage *ComponentsPage, componentName string) *ClusterComponent {
	count, _ := componentsPage.ClusterComponentsList.Count()
	for i := 0; i < count; i++ {
		listItemNode := componentsPage.ClusterComponentsList.At(i)
		name, _ := listItemNode.FindByClass("cluster-component-name").Text()
		statusNode := listItemNode.FindByXPath(`*[@data-status]`)
		if name == componentName {
			return &ClusterComponent{Name: name, StatusNode: statusNode}
		}
	}
	return &ClusterComponent{}
}
