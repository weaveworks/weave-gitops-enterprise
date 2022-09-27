package pages

import (
	"fmt"
	"time"

	gomega "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

func GetWindowName(webDriver *agouti.Page) string {
	var result string
	gomega.Expect(webDriver.RunScript(`return window.name`, map[string]interface{}{}, &result)).ShouldNot(gomega.HaveOccurred())
	return result
}

func SetWindowName(webDriver *agouti.Page, windowName string) {
	var result interface{}
	gomega.Expect(webDriver.RunScript(fmt.Sprintf(`window.name="%s"`, windowName), map[string]interface{}{}, &result)).ShouldNot(gomega.HaveOccurred())
}

func ElementExist(element *agouti.Selection, timeOutSec ...int) bool {
	timeout := 5
	if len(timeOutSec) > 0 {
		timeout = timeOutSec[0]
	}
	time.Sleep(500 * time.Millisecond) // Half secod delays to beign for stability
	for i := 1; i < timeout; i++ {
		if count, _ := element.Count(); count == 1 {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func ScrollWindow(webDriver *agouti.Page, xOffSet int, yOffSet int) {
	// script := fmt.Sprintf(`var elmnt = document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue; elmnt.scrollIntoView();`, xpath)

	script := fmt.Sprintf(`window.scrollTo(%d, %d)`, xOffSet, yOffSet)
	var result interface{}
	gomega.Expect(webDriver.RunScript(script, map[string]interface{}{}, &result)).ShouldNot(gomega.HaveOccurred())
}

func OpenWindowInBg(webDriver *agouti.Page, url string, windowName string) {
	currentWindow, err := webDriver.Session().GetWindow()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get current/active window")

	script := fmt.Sprintf(`window.open('%s', '%s')`, url, windowName)
	var result interface{}
	gomega.Expect(webDriver.RunScript(script, map[string]interface{}{}, &result)).ShouldNot(gomega.HaveOccurred(), "Failed to execute java script to open new window")
	gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch back to old window")
}

func CloseWindow(webDriver *agouti.Page, windowName string) {
	currentWindow, err := webDriver.Session().GetWindow()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get current/active window")
	gomega.Expect(webDriver.SwitchToWindow(windowName)).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to switch to %s window", windowName))
	gomega.Expect(webDriver.CloseWindow()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to close %s window", windowName))
	gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch back to old window")
}

func ClearFieldValue(field *agouti.Selection) {
	val, _ := field.Attribute("value")
	for i := 0; i < len(val); i++ {
		gomega.Expect(field.SendKeys("\uE003")).To(gomega.Succeed())
	}
}

func ClickElement(webDriver *agouti.Page, element *agouti.Selection, xOffset, yOffset int) error {
	gomega.Expect(element.MouseToElement()).Should(gomega.Succeed(), "Failed to move mouse to element")
	gomega.Expect(webDriver.MoveMouseBy(xOffset, yOffset)).Should(gomega.Succeed(), fmt.Sprintf("Failed to move mouse by offset (%d, %d)", xOffset, yOffset))
	return webDriver.Click(agouti.SingleClick, agouti.LeftButton)
}
