package acceptance

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("Workspaces", func() {

	BeforeEach(func() {
		var err error
		if webDriver == nil {

			webDriver, err = agouti.NewPage(seleniumServiceUrl, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"chromeOptions": map[string][]string{
					"args": {
						//"--headless", //Uncomment to run headless
						"--disable-gpu",
						"--no-sandbox",
					}}}))
			Expect(err).NotTo(HaveOccurred())

			// Make the page bigger so we can see all the things in the screenshots
			err = webDriver.Size(1440, 3000)
			Expect(err).NotTo(HaveOccurred())
		}

		workspacesUrl := wkpUrl + "/workspaces"
		By("When I navigate to WKP dashboard", func() {
			Expect(webDriver.Navigate(workspacesUrl)).To(Succeed())
		})
	})

	AfterEach(func() {
		TakeNextScreenshot()
	})

	It("Verify WKP Dashboard Page Structure", func() {
		var expectedWKPTitle = "WKP · Workspaces"
		Eventually(webDriver).Should(HaveTitle(expectedWKPTitle))
	})

	It("Should list the workspaces", func() {
		workspaces := webDriver.All(".MuiTableBody-root tr")
		Eventually(workspaces).Should(HaveCount(0))
	})

	It("Should show the workspaces dialog when clicked", func() {
		workspacesHeader := webDriver.First(".workspaces-header")
		Eventually(workspacesHeader).Should(BeFound())
		button := workspacesHeader.First("button")

		err := button.Click()
		Expect(err).NotTo(HaveOccurred())

		workspaceDialog := webDriver.First(".create-workspace-dialog")
		Eventually(workspaceDialog).Should(BeFound())
		TakeNextScreenshot()

		selectTeamRow := workspaceDialog.First(".create-workspace-dialog-team-row")
		Eventually(selectTeamRow).Should(BeFound())
		// FIXME: brittle, will break
		dropDownInput := selectTeamRow.First("[title=devs]")
		Eventually(dropDownInput).Should(BeFound())

		createButton := workspaceDialog.FirstByButton("Create workspace")
		Eventually(createButton).Should(BeFound())
		err = createButton.Click()
		Expect(err).NotTo(HaveOccurred())

		TakeNextScreenshot()

		workspaces := webDriver.All(".MuiTableBody-root tr")
		Eventually(workspaces, 60*time.Second).Should(HaveCount(1))

		workspace := webDriver.First(".MuiTableBody-root tr td")
		name, err := workspace.Text()
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("devs-workspace"))
	})
})
