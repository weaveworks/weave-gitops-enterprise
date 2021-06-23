package pages

import (
	"github.com/sclevine/agouti"
)

//Header webDriver elements
type CreateCluster struct {
	CreateHeader *agouti.Selection
}

//CreateCluster initialises the webDriver object
func GetCreateClusterPage(webDriver *agouti.Page) *CreateCluster {
	clusterPage := CreateCluster{
		CreateHeader: webDriver.Find(`.count-header`),
	}

	return &clusterPage
}
