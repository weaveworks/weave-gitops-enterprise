package pages

import (
	"fmt"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

// NavbarwebDriver webDriver elements
type NavbarwebDriver struct {
	Title        *agouti.Selection
	Clusters     *agouti.Selection
	Templates    *agouti.Selection
	Applications *agouti.Selection
	Policies     *agouti.Selection
	Violations   *agouti.Selection
	Workspaces   *agouti.Selection
}

// NavbarwebDriver initialises the webDriver object
func Navbar(webDriver *agouti.Page) *NavbarwebDriver {
	navbar := NavbarwebDriver{
		Title:        webDriver.Find(`nav div[title="Home"]`),
		Clusters:     webDriver.Find(`nav .nav-items a[href="/clusters"]`),
		Templates:    webDriver.Find(`nav .nav-items a[href="/templates"]`),
		Applications: webDriver.Find(`nav .nav-items a[href="/applications"]`),
		Policies:     webDriver.Find(`nav .nav-items a[href="/policies"]`),
		Violations:   webDriver.Find(`nav .nav-items a[href="/clusters/violations"]`),
		Workspaces:   webDriver.Find(`nav .nav-items a[href="/workspaces"]`),
	}

	return &navbar
}

func NavigateToPage(webDriver *agouti.Page, page string) {
	gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
	navbarPage := Navbar(webDriver)

	ginkgo.By(fmt.Sprintf("When I click %s from Navbar", page), func() {
		switch page {
		case "Clusters":
			gomega.Eventually(navbarPage.Clusters.Click).Should(gomega.Succeed())
		case "Templates":
			gomega.Eventually(navbarPage.Templates.Click).Should(gomega.Succeed())
		case "Applications":
			gomega.Eventually(navbarPage.Applications.Click).Should(gomega.Succeed())
		case "Policies":
			gomega.Eventually(navbarPage.Policies.Click).Should(gomega.Succeed())
		case "Violations":
			gomega.Eventually(navbarPage.Violations.Click).Should(gomega.Succeed())
		case "Workspaces":
			gomega.Eventually(navbarPage.Workspaces.Click).Should(gomega.Succeed())
		}
	})
}
