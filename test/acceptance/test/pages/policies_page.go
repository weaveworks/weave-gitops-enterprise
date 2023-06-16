package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type PoliciesPage struct {
	PolicyHeader     *agouti.Selection
	PolicyHeaderLink *agouti.Selection
	PoliciesList     *agouti.Selection
	AlertError       *agouti.Selection
}

type PolicyInformation struct {
	Name                *agouti.Selection
	Category            *agouti.Selection
	AuditMode           *agouti.Selection
	AuditModeIcon       *agouti.Selection
	EnforceMode         *agouti.Selection
	EnforceModeIcon     *agouti.Selection
	AuditModeNoneIcon   *agouti.Selection
	EnforceModeNoneIcon *agouti.Selection
	Tenant              *agouti.Selection
	Severity            *agouti.Selection
	Cluster             *agouti.Selection
	Age                 *agouti.Selection
}

type PolicyDetailPage struct {
	Header          *agouti.Selection
	ID              *agouti.Selection
	ClusterName     *agouti.Selection
	Tags            *agouti.MultiSelection
	Severity        *agouti.Selection
	Category        *agouti.Selection
	Mode            *agouti.Selection
	TargetedK8sKind *agouti.MultiSelection
	Description     *agouti.Selection
	HowToSolve      *agouti.Selection
	Code            *agouti.Selection
	Parameters      *agouti.MultiSelection
}

type PolicyParametersDetail struct {
	Name     *agouti.Selection
	Type     *agouti.Selection
	Value    *agouti.Selection
	Required *agouti.Selection
}

func (p PoliciesPage) FindPolicyInList(policyName string) *PolicyInformation {
	policy := p.PoliciesList.FindByXPath(fmt.Sprintf(`//tr[.//a[.="%s"]]`, policyName))
	return &PolicyInformation{
		Name:                policy.FindByXPath(`td[1]//a`),
		Category:            policy.FindByXPath(`td[2]`),
		AuditMode:           policy.FindByXPath(`td[3]`),
		AuditModeIcon:       policy.Find(`span[title="Audit"]`),
		EnforceMode:         policy.FindByXPath(`td[4]`),
		EnforceModeIcon:     policy.Find(`span[title='Enforce']`),
		AuditModeNoneIcon:   policy.FindByXPath(`td[3]/span/div/span[text()='-']`),
		EnforceModeNoneIcon: policy.FindByXPath(`td[4]/span/div/span[text()='-']`),
		Tenant:              policy.FindByXPath(`td[5]`),
		Severity:            policy.FindByXPath(`td[6]`),
		Cluster:             policy.FindByXPath(`td[7]`),
		Age:                 policy.FindByXPath(`td[8]`),
	}
}

type ParameterFields struct {
	Name     *agouti.Selection
	Type     *agouti.Selection
	Value    *agouti.Selection
	Required *agouti.Selection
}

func (p PoliciesPage) CountPolicies() int {
	policies := p.PoliciesList.All("tr")
	count, _ := policies.Count()
	return count
}

func GetPoliciesPage(webDriver *agouti.Page) *PoliciesPage {
	policyPage := PoliciesPage{
		PolicyHeader:     webDriver.Find(`span[title="Policies"]`),
		PolicyHeaderLink: webDriver.Find(`div[role="heading"] a[href="/policies"]`),
		PoliciesList:     webDriver.First(`table tbody`),
		AlertError:       webDriver.Find(`#alert-list-errors`),
	}
	return &policyPage
}

func GetPolicyDetailPage(webDriver *agouti.Page) *PolicyDetailPage {
	return &PolicyDetailPage{
		Header:          webDriver.Find(`div[class*=Page__TopToolBar] span[class*=Breadcrumbs]`),
		ID:              webDriver.Find(`div[data-testid="Policy ID"]`),
		ClusterName:     webDriver.Find(`div[data-testid="Cluster"]`),
		Tags:            webDriver.All(`div[data-testid="Tags"] span`),
		Severity:        webDriver.Find(`div[data-testid="Severity"]`),
		Category:        webDriver.Find(`div[data-testid="Category"]`),
		Mode:            webDriver.Find(`div[data-testid="Mode"]`),
		TargetedK8sKind: webDriver.All(`div[data-testid="Targeted K8s Kind"] span`),
		Description:     webDriver.FindByXPath(`//div[text()="Description:"]/following-sibling::*[1]`),
		HowToSolve:      webDriver.FindByXPath(`//div[text()="How to solve:"]/following-sibling::*[1]`),
		Code:            webDriver.FindByXPath(`//div[text()="Policy Code:"]/following-sibling::*[1]`),
		Parameters:      webDriver.AllByXPath(`//div/*[text()="Parameters Definition:"]/following-sibling::*`),
	}
}

func (p PolicyDetailPage) GetTags() []string {
	tags := []string{}
	tCount, _ := p.Tags.Count()

	for i := 0; i < tCount; i++ {
		tag, _ := p.Tags.At(i).Text()
		tags = append(tags, tag)
	}
	return tags
}

func (p PolicyDetailPage) GetTargetedK8sKind() []string {
	k8sKinds := []string{}
	kCount, _ := p.TargetedK8sKind.Count()

	for i := 0; i < kCount; i++ {
		kind, _ := p.TargetedK8sKind.At(i).Text()
		k8sKinds = append(k8sKinds, kind)
	}
	return k8sKinds
}

func (p PolicyDetailPage) GetParameter(parameterName string) *ParameterFields {
	pCount, _ := p.Parameters.Count()
	parameterFields := ParameterFields{}

	for i := 0; i < pCount; i++ {
		if pName, _ := p.Parameters.At(i).FindByXPath(`div[1]/span[2]`).Text(); pName == parameterName {
			parameterFields = ParameterFields{
				Name:     p.Parameters.At(i).FindByXPath(`div[1]/span[2]`),
				Type:     p.Parameters.At(i).FindByXPath(`div[2]/span[2]`),
				Value:    p.Parameters.At(i).FindByXPath(`div[3]/span[2]`),
				Required: p.Parameters.At(i).FindByXPath(`div[4]/span[2]`),
			}
		}
	}
	return &parameterFields
}
