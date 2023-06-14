package pages

import (
	"fmt"

	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/matchers"
)

// Header webDriver elements
type CreateCluster struct {
	CreateHeader *agouti.Selection
	// TemplateName   *agouti.Selection
	Credentials     *agouti.Selection
	TemplateSection *agouti.MultiSelection
	ProfileList     *agouti.Selection
	PreviewPR       *agouti.Selection
	AddApplication  *agouti.Selection
}

type ProfileInformation struct {
	Checkbox  *agouti.Selection
	Name      *agouti.Selection
	Layer     *agouti.Selection
	Version   *agouti.Selection
	Namespace *agouti.Selection
	Values    *agouti.Selection
}

type FormField struct {
	Label   *agouti.Selection
	Field   *agouti.Selection
	Focused *agouti.Selection
	ListBox *agouti.Selection
}

type ValuesYaml struct {
	Title    *agouti.Selection
	Cancel   *agouti.Selection
	Save     *agouti.Selection
	TextArea *agouti.Selection
}

type CostEstimation struct {
	Label    *agouti.Selection
	Price    *agouti.Selection
	Message  *agouti.Selection
	Estimate *agouti.Selection
}

func GetCreateClusterPage(webDriver *agouti.Page) *CreateCluster {
	clusterPage := CreateCluster{
		CreateHeader: webDriver.Find(`.count-header`),
		// TemplateName:   webDriver.FindByXPath(`//*/div[text()="Create new resource with template"]/following-sibling::text()`),
		Credentials:     webDriver.Find(`.credentials [role="button"]`),
		TemplateSection: webDriver.AllByXPath(`//div[contains(@class, "form-group field field-object")]/child::div`),
		ProfileList:     webDriver.Find(`.profiles-table table tbody`),
		PreviewPR:       webDriver.FindByButton("PREVIEW PR"),
		AddApplication:  webDriver.FindByButton("ADD AN APPLICATION"),
	}

	return &clusterPage
}

func (c CreateCluster) GetTemplateParameter(webdriver *agouti.Page, name string) FormField {
	gomega.Eventually(webdriver.FindByID(fmt.Sprintf(`%s-group`, name))).Should(matchers.BeFound())
	param := webdriver.FindByID(fmt.Sprintf(`%s-group`, name))

	return FormField{
		Label:   param.Find(`label`),
		Field:   param.Find(`input`),
		Focused: param.Find(`div.Mui-focused`),
		ListBox: param.Find(`div[role="button"][aria-haspopup="listbox"]`),
	}
}

func GetValuesYaml(webDriver *agouti.Page) ValuesYaml {
	return ValuesYaml{
		Title:    webDriver.Find(`div[class^=MuiDialogTitle-root] h5`),
		Cancel:   webDriver.Find(`div[class^=MuiDialogTitle-root] button`),
		TextArea: webDriver.FindByXPath(`//div[contains(@class, "MuiDialogContent-root")]/textarea[1]`),
		Save:     webDriver.Find(`button#edit-yaml`),
	}
}

func (c CreateCluster) GetProfileInList(profileName string) *ProfileInformation {
	cluster := c.ProfileList.FindByXPath(fmt.Sprintf(`//span[@data-profile-name="%s"]/ancestor::tr`, profileName))
	return &ProfileInformation{
		Checkbox:  cluster.FindByXPath(`td[1]//input`),
		Name:      cluster.FindByXPath(`td[2]`),
		Layer:     cluster.FindByXPath(`td[3]`),
		Version:   cluster.FindByXPath(`td//div[contains(@class, "profile-version")]`),
		Namespace: cluster.FindByXPath(`td//div[contains(@class, "profile-namespace")]//input`),
		Values:    cluster.FindByXPath(`td//div[contains(@class, "profile-version")]/following-sibling::button`),
	}
}

func GetCredentials(webDriver *agouti.Page) *agouti.MultiSelection {
	return webDriver.All(`div[role*=presentation] li[class*=MuiListItem-root]`)
}

func GetCredential(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`li[class*=MuiListItem-root][data-value="%s"]`, value))
}

func GetOption(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`li[data-value="%s"]`, value))
}

func GetCostEstimation(webDriver *agouti.Page) *CostEstimation {
	return &CostEstimation{
		Label:    webDriver.FindByXPath(`//h2[.="Cost Estimation"]`),
		Price:    webDriver.FindByXPath(`//div[.="Monthly Cost:"]/following-sibling::div`),
		Message:  webDriver.FindByXPath(`//h2[.="Cost Estimation"]/following-sibling::div//div[contains(@class, "message")]`),
		Estimate: webDriver.FindByID(`get-estimation`),
	}
}
