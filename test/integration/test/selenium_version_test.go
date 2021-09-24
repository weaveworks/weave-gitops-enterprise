package test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

const maxRetries = 10
const retrySleepInterval = 1

func TestVersionInUI(t *testing.T) {
	if os.Getenv("SELENIUM_DEBUG") == "true" {
		selenium.SetDebug(true)
	}

	// Connect to the WebDriver instance running remotely
	caps := selenium.Capabilities{"browserName": "chrome"}
	// Selenium on circleci is exposed at localhost:4444 by default
	wd, err := selenium.NewRemote(caps, "http://localhost:4444/wd/hub")
	if err != nil {
		t.Error("could not make selenium remote at localhost:4444\nerr: ", err)
	}
	defer func() {
		_ = wd.Quit()
	}()

	if err := wd.Get("http://localhost:8090"); err != nil {
		t.Error("could not get WKP UI at localhost:8090\nerr: ", err)
	}

	// Wait for WKP UI to load and get the version
	time.Sleep(time.Duration(5 * time.Second))

	pageSource, err := wd.PageSource()
	if err != nil {
		t.Error("could not get WKP UI page source\nerr: ", err)
	}
	_ = ioutil.WriteFile("/tmp/workspace/wkp-ui-page-source", []byte(pageSource), 0644)

	var output string
	retries := 0
	correctVersion := false
	for retries < maxRetries {
		versionElement, err := wd.FindElement(selenium.ByCSSSelector, "#wkp-ui-cluster-version")
		if err != nil {
			t.Log("did not find element #wkp-ui-cluster-version\nerr: ", err)
			retries++
			time.Sleep(time.Duration(retrySleepInterval * time.Second))
			continue
		}

		output, err = versionElement.GetAttribute("innerText")
		if err != nil {
			t.Error("could not get innerText of version element #wkp-ui-cluster-version\nerr: ", err)
			break
		}

		t.Log("WKP UI showing version: ", output)
		t.Logf("should show version: v%s\n", os.Getenv("CLUSTER_VERSIONS"))
		// if version is not correct, retry
		if output != "v"+os.Getenv("CLUSTER_VERSIONS") {
			retries++
			time.Sleep(time.Duration(retrySleepInterval * time.Second))
			continue
		}
		correctVersion = true
		break
	}
	// Take a screenshot of the rendered UI
	screenShot, err := wd.Screenshot()
	if err != nil {
		t.Error("failed to get screenshot of WKP UI\nerr: ", err)
	}
	_ = ioutil.WriteFile("/tmp/workspace/wkp-ui-screenshot.png", screenShot, 0644)

	if !correctVersion {
		t.Error("WKP UI is not showing the correct version")
	}
}
