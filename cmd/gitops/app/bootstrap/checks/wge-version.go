package checks

import (
	"fmt"
)

func CheckWgeVersion() {
	VERSIONS := []string{"v0.29.1", "v0.29.0", "v0.28.0"}
	versionSelectorPrompt := promptContent{
		"",
		"Please select a version for WGE to be installed",
	}
	selectedVersion := promptGetSelect(versionSelectorPrompt, VERSIONS)

	fmt.Printf("Installing Weave Gitops Enterprise ... %s\n", selectedVersion)
}
