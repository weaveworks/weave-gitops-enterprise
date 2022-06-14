package pages

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

//NavbarwebDriver webDriver elements
type NavbarwebDriver struct {
	Title        *agouti.Selection
	Clusters     *agouti.Selection
	Templates    *agouti.Selection
	Applications *agouti.Selection
	Policies     *agouti.Selection
}

//NavbarwebDriver initialises the webDriver object
func Navbar(webDriver *agouti.Page) *NavbarwebDriver {
	navbar := NavbarwebDriver{
		Title:        webDriver.Find(`nav a[title="Home"]`),
		Clusters:     webDriver.Find(`nav a[href="/clusters"]`),
		Templates:    webDriver.Find(`nav a[href="/clusters/templates"]`),
		Applications: webDriver.Find(`nav a[href="/applications"]`),
		Policies:     webDriver.Find(`nav a[href="/policies"]`),
	}

	return &navbar
}

func NavigateToPage(webDriver *agouti.Page, page string) {
	Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
	navbarPage := Navbar(webDriver)

	By(fmt.Sprintf("When I click %s from Navbar", page), func() {
		switch page {
		case "Clusters":
			Eventually(navbarPage.Clusters.Click).Should(Succeed())
		case "Templates":
			Eventually(navbarPage.Templates.Click).Should(Succeed())
		case "Applications":
			Eventually(navbarPage.Applications.Click).Should(Succeed())
		case "Policies":
			Eventually(navbarPage.Policies.Click).Should(Succeed())
		}
	})
}
