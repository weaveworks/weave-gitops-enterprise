package pages

import (
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

//Header webDriver elements
type CreateCluster struct {
	CreateHeader *agouti.Selection
	// TemplateName   *agouti.Selection
	Credentials     *agouti.Selection
	TemplateSection *agouti.MultiSelection
	ProfileList     *agouti.Selection
	PreviewPR       *agouti.Selection
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
	ListBox *agouti.Selection
}

type TemplateSection struct {
	Name   *agouti.Selection
	Fields []FormField
}

type ValuesYaml struct {
	Title    *agouti.Selection
	Cancel   *agouti.Selection
	Save     *agouti.Selection
	TextArea *agouti.Selection
}

type Preview struct {
	Title *agouti.Selection
	Text  *agouti.Selection
	Close *agouti.Selection
}

type GitOps struct {
	GitOpsLabel    *agouti.Selection
	GitOpsFields   []FormField
	GitCredentials *agouti.Selection
	CreatePR       *agouti.Selection
	SuccessBar     *agouti.Selection
	PRLinkBar      *agouti.Selection
	ErrorBar       *agouti.Selection
}

//CreateCluster initialises the webDriver object
func GetCreateClusterPage(webDriver *agouti.Page) *CreateCluster {
	clusterPage := CreateCluster{
		CreateHeader: webDriver.Find(`.count-header`),
		// TemplateName:   webDriver.FindByXPath(`//*/div[text()="Create new cluster with template"]/following-sibling::text()`),
		Credentials:     webDriver.Find(`.credentials [role="button"]`),
		TemplateSection: webDriver.AllByXPath(`//div[contains(@class, "form-group field field-object")]/child::div`),
		ProfileList:     webDriver.Find(`.profiles-table table tbody`),
		PreviewPR:       webDriver.FindByButton("PREVIEW PR"),
	}

	return &clusterPage
}

// This function waits for Create emplate page to load completely
func (c CreateCluster) WaitForPageToLoad(webDriver *agouti.Page) {
	// Credentials dropdown takes a while to populate
	Eventually(webDriver.Find(`.credentials [role="button"][aria-disabled="true"]`),
		30*time.Second).ShouldNot(BeFound())
	// With the introduction of profiles, UI takes long time to be fully rendered, UI refreshes once all the profiles valus are read and populated
	// This delay refresh sometimes cause tests to fail select elements
	time.Sleep(2 * time.Second)
}

func (c CreateCluster) GetTemplateSection(webdriver *agouti.Page, sectionName string) TemplateSection {
	paramSection := fmt.Sprintf(`div[data-name="%s"]`, sectionName)
	Eventually(webdriver.Find(paramSection)).Should(BeFound())
	section := webdriver.Find(paramSection)
	name := section.Find(".section-name")
	fields := section.All(".step-child")

	formFields := []FormField{}
	fCnt, _ := fields.Count()

	for i := 0; i < fCnt; i++ {
		formFields = append(formFields, FormField{
			Label:   fields.At(i).Find(`label`),
			Field:   fields.At(i).Find(`input`),
			ListBox: fields.At(i).Find(`div[role="button"][aria-haspopup="listbox"]`),
		})
	}

	return TemplateSection{
		Name:   name,
		Fields: formFields,
	}
}

func (c CreateCluster) GetTemplateParameter(webdriver *agouti.Page, name string) FormField {
	Eventually(webdriver.FindByID(fmt.Sprintf(`%s-group`, name))).Should(BeFound())
	param := webdriver.FindByID(fmt.Sprintf(`%s-group`, name))

	return FormField{
		Label:   param.Find(`label`),
		Field:   param.Find(`input`),
		ListBox: param.Find(`div[role="button"][aria-haspopup="listbox"]`),
	}
}

func GetValuesYaml(webDriver *agouti.Page) ValuesYaml {
	Eventually(webDriver.Find(`div[class^=MuiDialogTitle-root]`)).Should(BeVisible())
	return ValuesYaml{
		Title:    webDriver.Find(`div[class^=MuiDialogTitle-root] > h5`),
		Cancel:   webDriver.Find(`div[class^=MuiDialogTitle-root] > button`),
		TextArea: webDriver.FindByXPath(`//div[contains(@class, "MuiDialogContent-root")]/textarea[1]`),
		Save:     webDriver.Find(`button#edit-yaml`),
	}
}

// FindProfileInList finds the profile with given name
func (c CreateCluster) FindProfileInList(profileName string) *ProfileInformation {
	cluster := c.ProfileList.FindByXPath(fmt.Sprintf(`//span[@data-profile-name="%s"]/ancestor::tr`, profileName))
	return &ProfileInformation{
		Checkbox:  cluster.FindByXPath(`td[1]//input`),
		Name:      cluster.FindByXPath(`td[2]`),
		Layer:     cluster.FindByXPath(`td[3]`),
		Version:   cluster.FindByXPath(`td[4]//div[contains(@class, "profile-version")]`),
		Namespace: cluster.FindByXPath(`td[4]//div[contains(@class, "profile-namespace")]//input`),
		Values:    cluster.FindByXPath(`td[4]//button`),
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

func GetPreview(webDriver *agouti.Page) Preview {
	return Preview{
		Title: webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  h5`),
		Text:  webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  textarea:first-child`),
		Close: webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  button`),
	}
}

func GetGitOps(webDriver *agouti.Page) GitOps {
	return GitOps{
		GitOpsLabel: webDriver.FindByXPath(`//h2[.="GitOps"]`),
		GitOpsFields: []FormField{
			{
				Label: webDriver.FindByLabel(`CREATE BRANCH`),
				Field: webDriver.FindByID(`CREATE BRANCH-input`),
			},
			{
				Label: webDriver.FindByLabel(`PULL REQUEST TITLE`),
				Field: webDriver.FindByID(`PULL REQUEST TITLE-input`),
			},
			{
				Label: webDriver.FindByLabel(`COMMIT MESSAGE`),
				Field: webDriver.FindByID(`COMMIT MESSAGE-input`),
			},
			{
				Label: webDriver.FindByLabel(`PULL REQUEST DESCRIPTION`),
				Field: webDriver.FindByID(`PULL REQUEST DESCRIPTION-input`),
			},
		},
		GitCredentials: webDriver.Find(`div.auth-message`),
		CreatePR:       webDriver.FindByButton(`CREATE PULL REQUEST`),
		SuccessBar:     webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "Success")]/parent::div`),
		PRLinkBar:      webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "PR created")]/parent::div`),
		ErrorBar:       webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "Error")]/parent::div`),
	}
}
