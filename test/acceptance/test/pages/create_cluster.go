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
	Credentials        *agouti.Selection
	TemplateSection    *agouti.MultiSelection
	ProfileSelect      *agouti.Selection
	ProfileSelectPopup *agouti.MultiSelection
	PreviewPR          *agouti.Selection
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

type Profile struct {
	Name    *agouti.Selection
	Version *agouti.Selection
	Layer   *agouti.Selection
	Values  *agouti.Selection
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
		Credentials:        webDriver.Find(`.credentials [role="button"]`),
		TemplateSection:    webDriver.AllByXPath(`//div[contains(@class, "form-group field field-object")]/child::div`),
		ProfileSelect:      webDriver.Find(`div.profiles-select > div`),
		ProfileSelectPopup: webDriver.All(`ul[role="listbox"] li`),
		PreviewPR:          webDriver.FindByButton("PREVIEW PR"),
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

func GetProfile(webDriver *agouti.Page, profileName string) Profile {
	p := webDriver.Find(fmt.Sprintf(`.profiles-list [data-profile-name="%s"]`, profileName))
	return Profile{
		Name:    p.Find(`.profile-name`),
		Version: p.Find(`.profile-version`),
		Layer:   p.Find(`.profile-layer > span + span`),
		Values:  p.Find(`button`),
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

func (c CreateCluster) SelectProfile(profileName string) *agouti.Selection {
	time.Sleep(2 * time.Second)
	pCount, _ := c.ProfileSelectPopup.Count()

	for i := 0; i < pCount; i++ {
		pName, _ := c.ProfileSelectPopup.At(i).Text()
		if profileName == pName {
			return c.ProfileSelectPopup.At(i)
		}
	}
	return nil
}

func DissmissProfilePopup(webDriver *agouti.Page) {
	Expect(webDriver.Find(`div[name=Profiles]`).DoubleClick()).To(Succeed())
}

func GetCredentials(webDriver *agouti.Page) *agouti.MultiSelection {
	return webDriver.All(`li[class*=MuiListItem-root]`)
}

func GetCredential(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`li[class*=MuiListItem-root][data-value="%s"]`, value))
}

func GetOption(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`li[data-value="%s"]`, value))
}

func GetPreview(webDriver *agouti.Page) Preview {
	Eventually(webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]`)).Should(BeVisible())
	return Preview{
		Title: webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  h5`),
		Text:  webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  textarea:first-child`),
		Close: webDriver.Find(`div[class*=MuiDialog-paper][role=dialog]  button`),
	}
}

func GetGitOps(webDriver *agouti.Page) GitOps {
	return GitOps{
		GitOpsLabel: webDriver.FindByName("GitOps"),
		GitOpsFields: []FormField{
			{
				Label: webDriver.FindByLabel(`Create branch`),
				Field: webDriver.FindByID(`Create branch-input`),
			},
			{
				Label: webDriver.FindByLabel(`Pull request title`),
				Field: webDriver.FindByID(`Pull request title-input`),
			},
			{
				Label: webDriver.FindByLabel(`Commit message`),
				Field: webDriver.FindByID(`Commit message-input`),
			},
		},
		GitCredentials: webDriver.Find(`div.auth-message`),
		CreatePR:       webDriver.FindByButton(`CREATE PULL REQUEST`),
		SuccessBar:     webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "Success")]/parent::div`),
		PRLinkBar:      webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "PR created")]/parent::div`),
		ErrorBar:       webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//*[contains(text(), "Error")]/parent::div`),
	}
}
