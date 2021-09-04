package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/entitlements/cmd"
)

func init() {
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		// Only log the warning severity or above.
		log.SetLevel(log.DebugLevel)
	} else if os.Getenv("LOG_LEVEL") == "WARN" || os.Getenv("LOG_LEVEL") == "" {
		// Only log the warning severity or above.
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	cmd.Execute()
}
