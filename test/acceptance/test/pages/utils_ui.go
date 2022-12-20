package pages

import (
	"fmt"
	"reflect"
	"time"

	gomega "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/api"
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

func CloseWindow(webDriver *agouti.Page, windowToClose interface{}) {
	currentWindow, err := webDriver.Session().GetWindow()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get current/active window")

	vType := reflect.TypeOf(windowToClose)
	if vType.Elem().Kind() == reflect.String {
		gomega.Expect(webDriver.SwitchToWindow(windowToClose.(string))).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to switch to %s window", windowToClose.(string)))
	} else if vType.Elem().Kind() == reflect.Struct {
		gomega.Expect(webDriver.Session().SetWindow(windowToClose.(*api.Window))).ShouldNot(gomega.HaveOccurred(), "Failed to switch back to old window")
	}

	gomega.Expect(webDriver.CloseWindow()).ShouldNot(gomega.HaveOccurred(), "Failed to close window")
	gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch back to old window")
}

func GetNextWindow(webDriver *agouti.Page) *api.Window {
	currentWindow, err := webDriver.Session().GetWindow()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get current/active window")

	gomega.Expect(webDriver.NextWindow()).ShouldNot(gomega.HaveOccurred(), "Failed to switch to next window")
	window, err := webDriver.Session().GetWindow()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get opened active window")

	gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch back to old window")

	return window
}

func CloseOtherWindows(webDriver *agouti.Page, enterpriseWindow *api.Window) {
	gomega.Expect(webDriver.Session().SetWindow(enterpriseWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to specified window")

	windows, err := webDriver.Session().GetWindows()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get Windows")
	for _, window := range windows {
		if window.ID != enterpriseWindow.ID {
			CloseWindow(webDriver, window)
		}
	}
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
