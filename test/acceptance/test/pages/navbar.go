package pages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

//NavbarwebDriver webDriver elements
type NavbarwebDriver struct {
	Title       *agouti.Selection
	Clusters    *agouti.Selection
	Templates   *agouti.Selection
	Alerts      *agouti.Selection
	Application *agouti.Selection
}

//NavbarwebDriver initialises the webDriver object
func Navbar(webDriver *agouti.Page) *NavbarwebDriver {
	navbar := NavbarwebDriver{
		Title:       webDriver.Find(`nav a[title="Home"]`),
		Clusters:    webDriver.Find(`nav a[href="/clusters"]`),
		Templates:   webDriver.Find(`nav a[href="/clusters/templates"]`),
		Alerts:      webDriver.Find(`nav a[href="/clusters/alerts"]`),
		Application: webDriver.Find(`nav a[href="/applications"]`),
	}

	return &navbar
}

func NavigateToPage(webDriver *agouti.Page, page string) {
	webDriver.Refresh()
	navbarPage := Navbar(webDriver)
	By(fmt.Sprintf("When I click %s from Navbar", page), func() {
		Eventually(navbarPage.Templates).Should(HaveText(page))
		Expect(navbarPage.Templates.Click()).To(Succeed())
	})
}
