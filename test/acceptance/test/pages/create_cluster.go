package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

//Header webDriver elements
type CreateCluster struct {
	CreateHeader *agouti.Selection
	// TemplateName   *agouti.Selection
	ClusterSection *agouti.Selection
	PreviewPR      *agouti.Selection
}

type FormFeild struct {
	Label *agouti.Selection
	Feild *agouti.Selection
}

type Preview struct {
	PreviewLabel *agouti.Selection
	PreviewText  *agouti.Selection
}

type GitOps struct {
	GitOpsLabel  *agouti.Selection
	GitOpsFeilds []FormFeild
	CreatePR     *agouti.Selection
}

//CreateCluster initialises the webDriver object
func GetCreateClusterPage(webDriver *agouti.Page) *CreateCluster {
	clusterPage := CreateCluster{
		CreateHeader: webDriver.Find(`.count-header`),
		// TemplateName:   webDriver.FindByXPath(`//*/div[text()="Create new cluster with template"]/following-sibling::text()`),
		ClusterSection: webDriver.FindByXPath(`//*/h5[text()="Cluster"]/parent::div/following-sibling::div`),
		PreviewPR:      webDriver.FindByButton("Preview PR"),
	}

	return &clusterPage
}

func (c CreateCluster) GetTemplateParameter(paramName string) FormFeild {
	return FormFeild{
		Label: c.ClusterSection.Find(fmt.Sprintf(`#root_%s-label`, paramName)),
		Feild: c.ClusterSection.Find(fmt.Sprintf(`#root_%s`, paramName)),
	}
}

func GetPreview(webDriver *agouti.Page) Preview {
	return Preview{
		PreviewLabel: webDriver.FindByXPath(`//*/span[text()="Preview & Commit"]`),
		PreviewText:  webDriver.FindByXPath(`//*/span[contains(., "Preview")]/parent::div/following-sibling::textarea`),
	}
}

func GetGitOps(webDriver *agouti.Page) GitOps {
	return GitOps{
		GitOpsLabel: webDriver.FindByXPath(`//*/span[text()="GitOps"]`),
		GitOpsFeilds: []FormFeild{
			{
				Label: webDriver.FindByLabel(`Create branch`),
				Feild: webDriver.FindByID(`Create branch-input`),
			},
			{
				Label: webDriver.FindByLabel(`Title pull request`),
				Feild: webDriver.FindByID(`Title pull request-input`),
			},
			{
				Label: webDriver.FindByLabel(`Commit message`),
				Feild: webDriver.FindByID(`Commit message-input`),
			},
		},
		CreatePR: webDriver.FindByButton(`Create Pull Request on GitHub`),
	}
}
