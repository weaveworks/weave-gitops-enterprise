package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

type PoliciesPage struct {
	PolicyHeader *agouti.Selection
	PolicyCount  *agouti.Selection
	PoliciesList *agouti.Selection
}

type PolicyInformation struct {
	Name     *agouti.Selection
	Category *agouti.Selection
	Severity *agouti.Selection
	Cluster  *agouti.Selection
	Age      *agouti.Selection
}

type PolicyDetailPage struct {
	Header          *agouti.Selection
	Title           *agouti.Selection
	ID              *agouti.Selection
	ClusterName     *agouti.Selection
	Tags            *agouti.MultiSelection
	Severity        *agouti.Selection
	Category        *agouti.Selection
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
		Name:     policy.FindByXPath(`td[1]`),
		Category: policy.FindByXPath(`td[2]`),
		Severity: policy.FindByXPath(`td[3]`),
		Cluster:  policy.FindByXPath(`td[4]`),
		Age:      policy.FindByXPath(`td[5]`),
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
		PolicyHeader: webDriver.Find(`div[role="heading"] a[href="/policies"]`),
		PolicyCount:  webDriver.Find(`.section-header-count`),
		PoliciesList: webDriver.First(`table tbody`),
	}
	return &policyPage
}

func GetPolicyDetailPage(webDriver *agouti.Page) *PolicyDetailPage {
	return &PolicyDetailPage{
		Header:          webDriver.FindByXPath(`//div[@role="heading"]/a[@href="/policies"]/parent::node()/parent::node()/following-sibling::div`),
		Title:           webDriver.First(`h2`),
		ID:              webDriver.FindByXPath(`//div[text()="Policy ID:"]/following-sibling::*[1]`),
		ClusterName:     webDriver.FindByXPath(`//div[text()="Cluster Name:"]/following-sibling::*[1]`),
		Tags:            webDriver.AllByXPath(`//div/*[text()="Tags:"]/following-sibling::*`),
		Severity:        webDriver.FindByXPath(`//div[text()="Severity:"]/following-sibling::*[1]`),
		Category:        webDriver.FindByXPath(`//div[text()="Category:"]/following-sibling::*[1]`),
		TargetedK8sKind: webDriver.AllByXPath(`//div[text()="Targeted K8s Kind:"]/following-sibling::*`),
		Description:     webDriver.FindByXPath(`//div[text()="Description:"]/following-sibling::*[1]`),
		HowToSolve:      webDriver.FindByXPath(`//div[text()="How to solve:"]/following-sibling::*[1]`),
		Code:            webDriver.FindByXPath(`//div[text()="Policy Code:"]/following-sibling::*[1]`),
		Parameters:      webDriver.AllByXPath(`//div/*[text()="Parameters Definition"]/following-sibling::*`),
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
